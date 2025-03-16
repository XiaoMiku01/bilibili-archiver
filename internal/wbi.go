package internal

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	cacheImgKey = "imgKey"
	cacheSubKey = "subKey"
)

var (
	_defaultMixinKeyEncTab = []int{
		46, 47, 18, 2, 53, 8, 23, 32, 15, 50, 10, 31, 58, 3, 45, 35, 27, 43, 5, 49,
		33, 9, 42, 19, 29, 28, 14, 39, 12, 38, 41, 13, 37, 48, 7, 16, 24, 55, 40,
		61, 26, 17, 0, 1, 60, 51, 30, 4, 22, 25, 54, 21, 56, 59, 6, 63, 57, 62, 11,
		36, 20, 34, 44, 52,
	}

	_defaultStorage = &MemoryStorage{
		data: make(map[string]any, 15),
	}
)

type Storage interface {
	Set(key string, value any)
	Get(key string) (v any, isSet bool)
}

type MemoryStorage struct {
	data map[string]any
	mu   sync.RWMutex
}

func (impl *MemoryStorage) Set(key string, value any) {
	impl.mu.Lock()
	defer impl.mu.Unlock()

	impl.data[key] = value
}

func (impl *MemoryStorage) Get(key string) (v any, isSet bool) {
	impl.mu.RLock()
	defer impl.mu.RUnlock()

	if v, isSet = impl.data[key]; isSet {
		return v, true
	}
	return nil, false
}

// WBI 签名实现
// 如果希望以登录的方式获取则使用 WithCookies or WithRawCookies 设置cookie
// 如果希望以未登录的方式获取 WithCookies(nil) 设置cookie为 nil 即可, 这是 Default 行为
//
//	!!! 使用 WBI 的接口 绝对不可以 set header Referer 会导致失败 !!!
//	!!! 大部分使用 WBI 的接口都需要 set header Cookie !!!
//
// see https://github.com/SocialSisterYi/bilibili-API-collect/blob/master/docs/misc/sign/wbi.md
type WBI struct {
	cookies        []*http.Cookie
	mixinKeyEncTab []int

	// updateCheckerInterval is the interval to check and update wbi keys
	// default is 60 minutes. so if lastInitTime + updateCheckerInterval < now, it will update wbi keys
	updateCheckerInterval time.Duration
	lastInitTime          time.Time
	storage               Storage
}

func NewDefaultWbi() *WBI {
	return &WBI{
		cookies:        nil,
		mixinKeyEncTab: _defaultMixinKeyEncTab,

		updateCheckerInterval: 60 * time.Minute,
		storage:               _defaultStorage,
	}
}

func (wbi *WBI) WithUpdateInterval(updateInterval time.Duration) *WBI {
	wbi.updateCheckerInterval = updateInterval
	return wbi
}

func (wbi *WBI) WithCookies(cookies []*http.Cookie) *WBI {
	wbi.cookies = cookies
	return wbi
}

func (wbi *WBI) WithRawCookies(rawCookies string) *WBI {
	header := http.Header{}
	header.Add("Cookie", rawCookies)
	req := http.Request{Header: header}

	wbi.cookies = req.Cookies()
	return wbi
}

func (wbi *WBI) WithMixinKeyEncTab(mixinKeyEncTab []int) *WBI {
	wbi.mixinKeyEncTab = mixinKeyEncTab
	return wbi
}

func (wbi *WBI) WithStorage(storage Storage) *WBI {
	wbi.storage = storage
	return wbi
}

func (wbi *WBI) GetKeys() (imgKey string, subKey string, err error) {
	imgKey, subKey = wbi.getKeys()

	return imgKey, subKey, nil
}

func (wbi *WBI) getKeys() (imgKey string, subKey string) {
	if v, isSet := wbi.storage.Get(cacheImgKey); isSet {
		imgKey = v.(string)
	}

	if v, isSet := wbi.storage.Get(cacheSubKey); isSet {
		subKey = v.(string)
	}

	return imgKey, subKey
}

func (wbi *WBI) SetKeys(imgKey, subKey string) {
	wbi.storage.Set(cacheImgKey, imgKey)
	wbi.storage.Set(cacheSubKey, subKey)
	wbi.lastInitTime = time.Now()
}

func (wbi *WBI) GetMixinKey() (string, error) {
	imgKey, subKey, err := wbi.GetKeys()
	if err != nil {
		return "", err
	}

	return wbi.GenerateMixinKey(imgKey + subKey), nil
}

func (wbi *WBI) GenerateMixinKey(orig string) string {
	var str strings.Builder
	for _, v := range wbi.mixinKeyEncTab {
		if v < len(orig) {
			str.WriteByte(orig[v])
		}
	}
	return str.String()[:32]
}

func (wbi *WBI) sanitizeString(s string) string {
	unwantedChars := []string{"!", "'", "(", ")", "*"}
	for _, char := range unwantedChars {
		s = strings.ReplaceAll(s, char, "")
	}
	return s
}

func (wbi *WBI) SignQuery(bf *BiliFrom, ts time.Time) *BiliFrom {
	payload := make(map[string]string, 10)
	query := bf.Get()
	for k := range query {
		payload[k] = fmt.Sprintf("%v", query[k])
	}

	newPayload, err := wbi.SignMap(payload, ts)
	if err != nil {
		return bf
	}

	newQuery := NewBiliFrom(map[string]any{})
	for k, v := range newPayload {
		newQuery.Set(k, v)
	}

	return newQuery
}

func (wbi *WBI) SignMap(payload map[string]string, ts time.Time) (newPayload map[string]string, err error) {
	newPayload = make(map[string]string, 10)
	for k, v := range payload {
		newPayload[k] = v
	}

	newPayload["wts"] = strconv.FormatInt(ts.Unix(), 10)

	// Sort keys
	keys := make([]string, 0, 10)
	for k := range newPayload {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	// Remove unwanted characters
	for k, v := range newPayload {
		v = wbi.sanitizeString(v)
		newPayload[k] = v
	}

	// Build URL parameters
	signQuery := url.Values{}
	for _, k := range keys {
		signQuery.Set(k, newPayload[k])
	}
	signQueryStr := signQuery.Encode()

	// Get mixin key
	mixinKey, err := wbi.GetMixinKey()
	if err != nil {
		return payload, err
	}

	// Calculate w_rid
	hash := md5.Sum([]byte(signQueryStr + mixinKey))
	newPayload["w_rid"] = hex.EncodeToString(hash[:])

	return newPayload, nil
}
