package internal

import (
	"crypto/md5"
	"path"
	"strings"

	"encoding/json"
	"fmt"

	"net/http"
	"net/url"
	"os"
	"sort"
	"time"

	"github.com/imroc/req/v3"
	"github.com/rs/zerolog/log"

	"google.golang.org/grpc"

	playapi "github.com/XiaoMiku01/bilibili-grpc-api-go/bilibili/app/playurl/v1"
	viewapi "github.com/XiaoMiku01/bilibili-grpc-api-go/bilibili/app/view/v1"
	dmapi "github.com/XiaoMiku01/bilibili-grpc-api-go/bilibili/community/service/dm/v1"
)

const (
	TV_APP_KEY = "4409e2ce8ffd12b8"
	TV_APP_SEC = "59b43e04ad6965f34319062b478f83dd"
)

type BiliFrom struct {
	data map[string]any
}

func NewBiliFrom(data map[string]any) *BiliFrom {
	return &BiliFrom{data: data}
}

func (bf *BiliFrom) Signature() {
	bf.data["appkey"] = TV_APP_KEY
	keys := make([]string, 0, len(bf.data))
	for k := range bf.data {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	query := ""
	for _, k := range keys {
		query += k + "=" + url.QueryEscape(fmt.Sprintf("%v", bf.data[k])) + "&"
	}
	query = query[:len(query)-1] + TV_APP_SEC
	hash := md5.Sum([]byte(query))
	bf.data["sign"] = fmt.Sprintf("%x", hash)
}

func (bf *BiliFrom) Get() map[string]any {
	return bf.data
}

func (bf *BiliFrom) Set(key string, value any) {
	bf.data[key] = value
}

func (bf *BiliFrom) String() string {
	return fmt.Sprintf("%v", bf.data)
}

func (ck *CookieInfoStruct) SaveToFile(cfName string) error {
	cookieInfoJson, err := json.MarshalIndent(ck, "", "  ")
	if err != nil {
		return err
	}
	cookieFile, err := os.Create(cfName)
	if err != nil {
		return err
	}
	defer cookieFile.Close()
	cookieFile.WriteString(string(cookieInfoJson))
	return nil
}

type BApiClient struct {
	client     *req.Client
	cookieFile string
	wbi        *WBI
	// gRPC相关
	grpcConn      *grpc.ClientConn
	accessKey     string
	dmClient      dmapi.DMClient
	playurlClient playapi.PlayURLClient
	viewClient    viewapi.ViewClient
}

// NewBApiClient 创建并初始化 BApiClient
func NewBApiClient() *BApiClient {
	// 初始化req.Client
	c := req.C().
		SetUserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.0.0 Safari/537.36 Edg/134.0.0.0").
		SetTimeout(5*time.Second).
		SetCommonErrorResult(&BiliErr{}).
		// 设置自动重试，最多重试3次
		SetCommonRetryCount(3).
		SetCommonRetryBackoffInterval(1*time.Second, 2*time.Second).
		SetCommonRetryCondition(func(resp *req.Response, err error) bool {
			// 如果有网络错误或其他HTTP错误，进行重试
			if err != nil {
				// 如果是B站API错误也进行重试
				if err, ok := err.(*BiliErr); ok {
					// -101 可能是未登录或需要重新登录
					if err.Code == -101 {
						log.Warn().Msg("cookie 失效，刷新 cookie")
						// 重新登录	TODO
						return false
					} else if err.Code == 86039 {
						return false // 忽略 扫码登录影响
					}
				}
				log.Warn().Stack().Err(err).Msg("请求失败, 进行重试")
				return true
			}

			return false
		}).
		OnAfterResponse(func(client *req.Client, resp *req.Response) error {
			// 解析响应
			var biliResp BiliResp
			if err := resp.UnmarshalJson(&biliResp); err != nil {
				return fmt.Errorf("解析响应失败: %w", err)
			}
			// 检查响应码
			if biliResp.Code != 0 {
				return &BiliErr{
					Code:    biliResp.Code,
					Message: biliResp.Message,
				}
			}
			// 如果响应码为0，使用Data字段重写结果
			if result := resp.SuccessResult(); result != nil {
				// 将 Data 重新序列化为 JSON
				dataBytes, err := json.Marshal(biliResp.Data)
				if err != nil {
					return fmt.Errorf("序列化数据失败: %w", err)
				}
				// 将 JSON 解析到结果对象中
				if err := json.Unmarshal(dataBytes, result); err != nil {
					return fmt.Errorf("解析数据失败: %w", err)
				}
			}
			return nil
		})

	return &BApiClient{
		client: c,
		wbi:    NewDefaultWbi(),
	}
}

func (ba *BApiClient) SetDev(log req.Logger) {
	ba.client = ba.client.DevMode().SetLogger(log)
}

func (ba *BApiClient) GET(api string, bf *BiliFrom, resuult any, wbi ...any) error {
	if bf != nil {
		if len(wbi) == 0 {
			bf.Signature()
		} else {
			bf = ba.wbi.SignQuery(bf, time.Now())
		}
	} else {
		bf = NewBiliFrom(map[string]any{})
	}
	_, err := ba.client.R().SetQueryParamsAnyType(bf.Get()).SetSuccessResult(resuult).Get(api)
	if err != nil {
		return err
	}
	return nil
}

func (ba *BApiClient) POST(api string, bf *BiliFrom, resuult any) error {
	if bf != nil {
		bf.Signature()
	} else {
		bf = NewBiliFrom(map[string]any{})
	}
	_, err := ba.client.R().SetFormDataAnyType(bf.Get()).SetSuccessResult(resuult).Post(api)
	if err != nil {
		return err
	}
	return nil
}

func (ba *BApiClient) SetCookieFile(cookieFile string) error {
	ba.cookieFile = cookieFile
	var cookieInfo CookieInfoStruct
	cf, err := os.ReadFile(cookieFile)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(cf, &cookieInfo); err != nil {
		return err
	}
	for _, cookie := range cookieInfo.CookieInfo.Cookies {
		ba.client = ba.client.SetCommonCookies(&http.Cookie{
			Name:  cookie.Name,
			Value: cookie.Value,
		})
	}
	ba.accessKey = cookieInfo.TokenInfo.AccessToken
	return nil
}

func (ba *BApiClient) GetQRCode() (QRCodeStruct, error) {
	api := "http://passport.bilibili.com/x/passport-tv-login/qrcode/auth_code"
	bf := NewBiliFrom(map[string]any{
		"local_id": 0,
		"ts":       time.Now().Unix(),
	})
	var result QRCodeStruct
	err := ba.POST(api, bf, &result)
	if err != nil {
		return QRCodeStruct{}, err
	}
	return result, nil
}

func (ba *BApiClient) VerifyQrCode(qrcode QRCodeStruct) (CookieInfoStruct, error) {

	api := "http://passport.bilibili.com/x/passport-tv-login/qrcode/poll"
	bf := NewBiliFrom(map[string]any{
		"auth_code": qrcode.AuthCode,
		"local_id":  0,
		"ts":        time.Now().Unix(),
	})
	var result CookieInfoStruct
	err := ba.POST(api, bf, &result)
	if err != nil {
		return CookieInfoStruct{}, err
	}
	return result, nil
}

func (ba *BApiClient) RefreshToken() (CookieInfoStruct, error) {
	api := "https://passport.bilibili.com/x/passport-login/oauth2/refresh_token"
	var cookieInfo CookieInfoStruct
	cf, err := os.ReadFile(ba.cookieFile)
	if err != nil {
		return CookieInfoStruct{}, err
	}
	if err := json.Unmarshal(cf, &cookieInfo); err != nil {
		return CookieInfoStruct{}, err
	}
	bf := NewBiliFrom(map[string]any{
		"access_token":  cookieInfo.TokenInfo.AccessToken,
		"refresh_token": cookieInfo.TokenInfo.RefreshToken,
		"actionKey":     "appkey",
		"ts":            time.Now().Unix(),
	})
	var result CookieInfoStruct
	err = ba.POST(api, bf, &result)
	if err != nil {
		return CookieInfoStruct{}, err
	}
	return result, nil
}

func (ba *BApiClient) GetUserInfo() (UserInfoStruct, error) {
	api := "https://api.bilibili.com/x/web-interface/nav"
	var result UserInfoStruct

	err := ba.GET(api, nil, &result)
	if err != nil {
		return UserInfoStruct{}, err
	}
	// set wbi
	ba.wbi.SetKeys(strings.Split(path.Base(result.WbiImg.ImgUrl), ".")[0], strings.Split(path.Base(result.WbiImg.SubUrl), ".")[0])
	return result, nil
}

func (ba *BApiClient) GetPlayURL(aid, cid int64) (PlayInfoStruct, error) {
	// ba.GetUserInfo() // 更新 wbi
	api := "https://api.bilibili.com/x/player/playurl"
	bf := NewBiliFrom(map[string]any{
		"avid":  aid,
		"cid":   cid,
		"fnval": 16 | 128,
		"fourk": 1,
	})
	var result PlayInfoStruct
	err := ba.GET(api, bf, &result, true)
	if err != nil {
		return PlayInfoStruct{}, err
	}
	return result, nil
}
func (ba *BApiClient) GetFavList(mid int) (FavListStruct, error) {
	api := "https://api.bilibili.com/x/v3/fav/folder/created/list-all"
	bf := NewBiliFrom(map[string]any{
		"up_mid": mid,
	})
	var result FavListStruct
	err := ba.GET(api, bf, &result)
	if err != nil {
		return FavListStruct{}, err
	}
	return result, nil
}

func (ba *BApiClient) GetFavMediaList(fid, pn int) (FavMediaListStruct, error) {
	api := "https://api.bilibili.com/x/v3/fav/resource/list"
	bf := NewBiliFrom(map[string]any{
		"media_id": fid,
		"order":    "mtime",
		"ps":       40,
		"pn":       pn,
		"type":     0,
		"tid":      0,
	})
	var result FavMediaListStruct
	err := ba.GET(api, bf, &result)
	if err != nil {
		return FavMediaListStruct{}, err
	}
	return result, nil
}

var BApi = NewBApiClient()

func CheckCookieFile(cfName string) (UserInfoStruct, bool) {
	BApi.SetCookieFile(cfName)
	var uf UserInfoStruct
	uf, err := BApi.GetUserInfo()
	if err != nil {
		log.Error().Err(err).Msg("获取用户信息失败")
		return UserInfoStruct{}, false
	}
	return uf, true
}

func RefreshToken(cfName string) {
	BApi.SetCookieFile(cfName)
	cookieInfo, err := BApi.RefreshToken()
	if err != nil {
		log.Fatal().Err(err).Msg("刷新 cookie 失败")
		// TODO 通知
		return
	}
	cookieInfo.SaveToFile(cfName)
	if uf, err := CheckCookieFile(cfName); err {
		log.Info().Msgf("%s UID: %v Cookie 刷新成功", uf.Uname, uf.Mid)
	}
}
