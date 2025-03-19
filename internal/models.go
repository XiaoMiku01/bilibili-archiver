package internal

import "fmt"

type BiliErr struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (be *BiliErr) Error() string {
	return fmt.Sprintf("code: %d, message: %s", be.Code, be.Message)
}

type BiliResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

type QRCodeStruct struct {
	AuthCode string `json:"auth_code"`
	Url      string `json:"url"`
}

type CookieInfoStruct struct {
	IsNew        bool   `json:"is_new"`
	Mid          int    `json:"mid"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenInfo    struct {
		Mid          int    `json:"mid"`
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
	} `json:"token_info"`
	CookieInfo struct {
		Cookies []struct {
			Name     string `json:"name"`
			Value    string `json:"value"`
			HTTPOnly int    `json:"http_only"`
			Expires  int    `json:"expires"`
			Secure   int    `json:"secure"`
		} `json:"cookies"`
		Domains []string `json:"domains"`
	} `json:"cookie_info"`
	Sso  []string `json:"sso"`
	Hint string   `json:"hint"`
}

type TokenInfoStruct struct {
	Mid         int64  `json:"mid"`
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	Refresh     bool   `json:"refresh"`
}

type UserInfoStruct struct {
	Uname  string `json:"uname"`
	Mid    int    `json:"mid"`
	WbiImg struct {
		ImgUrl string `json:"img_url"`
		SubUrl string `json:"sub_url"`
	} `json:"wbi_img"`
}

type FavListStruct struct {
	Count int `json:"count"`
	List  []struct {
		ID         int    `json:"id"`
		Fid        int    `json:"fid"`
		Mid        int64  `json:"mid"`
		Attr       int    `json:"attr"`
		Title      string `json:"title"`
		FavState   int    `json:"fav_state"`
		MediaCount int    `json:"media_count"`
	} `json:"list"`
}

type FavMediaListStruct struct {
	Info struct {
		ID    int    `json:"id"`
		Fid   int    `json:"fid"`
		Mid   int64  `json:"mid"`
		Attr  int    `json:"attr"`
		Title string `json:"title"`
		Cover string `json:"cover"`
		Upper struct {
			Mid       int64  `json:"mid"`
			Name      string `json:"name"`
			Face      string `json:"face"`
			Followed  bool   `json:"followed"`
			VipType   int    `json:"vip_type"`
			VipStatue int    `json:"vip_statue"`
		} `json:"upper"`
		CoverType int `json:"cover_type"`
		CntInfo   struct {
			Collect int `json:"collect"`
			Play    int `json:"play"`
			ThumbUp int `json:"thumb_up"`
			Share   int `json:"share"`
		} `json:"cnt_info"`
		Type       int    `json:"type"`
		Intro      string `json:"intro"`
		Ctime      int    `json:"ctime"`
		Mtime      int    `json:"mtime"`
		State      int    `json:"state"`
		FavState   int    `json:"fav_state"`
		LikeState  int    `json:"like_state"`
		MediaCount int    `json:"media_count"`
		IsTop      bool   `json:"is_top"`
	} `json:"info"`
	Medias  []FavMediaStruct `json:"medias"`
	HasMore bool             `json:"has_more"`
	TTL     int              `json:"ttl"`
}

type FavMediaStruct struct {
	ID       int64  `json:"id"`
	Type     int    `json:"type"`
	Title    string `json:"title"`
	Cover    string `json:"cover"`
	Intro    string `json:"intro"`
	Page     int    `json:"page"`
	Duration int    `json:"duration"`
	Upper    struct {
		Mid  int    `json:"mid"`
		Name string `json:"name"`
		Face string `json:"face"`
	} `json:"upper"`
	Attr    int `json:"attr"`
	CntInfo struct {
		Collect    int    `json:"collect"`
		Play       int    `json:"play"`
		Danmaku    int    `json:"danmaku"`
		Vt         int    `json:"vt"`
		PlaySwitch int    `json:"play_switch"`
		Reply      int    `json:"reply"`
		ViewText1  string `json:"view_text_1"`
	} `json:"cnt_info"`
	Link    string `json:"link"`
	Ctime   int    `json:"ctime"`
	Pubtime int    `json:"pubtime"`
	FavTime int    `json:"fav_time"`
	BvID    string `json:"bv_id"`
	Bvid    string `json:"bvid"`
	Season  any    `json:"season"`
	Ogv     any    `json:"ogv"`
	Ugc     struct {
		FirstCid int64 `json:"first_cid"`
	} `json:"ugc"`
	MediaListLink string `json:"media_list_link"`
}

type PlayInfoStruct struct {
	From              string   `json:"from"`
	Result            string   `json:"result"`
	Message           string   `json:"message"`
	Quality           int      `json:"quality"`
	Format            string   `json:"format"`
	Timelength        int      `json:"timelength"`
	AcceptFormat      string   `json:"accept_format"`
	AcceptDescription []string `json:"accept_description"`
	AcceptQuality     []int    `json:"accept_quality"`
	VideoCodecid      int      `json:"video_codecid"`
	SeekParam         string   `json:"seek_param"`
	SeekType          string   `json:"seek_type"`
	Durl              any      `json:"durl"`
	Dash              struct {
		Duration      int     `json:"duration"`
		MinBufferTime float64 `json:"min_buffer_time"`
		Video         []struct {
			ID           int      `json:"id"`
			BaseURL      string   `json:"base_url"`
			BackupURL    []string `json:"backup_url"`
			Bandwidth    int      `json:"bandwidth"`
			MimeType     string   `json:"mime_type"`
			Codecs       string   `json:"codecs"`
			Width        int      `json:"width"`
			Height       int      `json:"height"`
			FrameRate    string   `json:"frame_rate"`
			Sar          string   `json:"sar"`
			StartWithSap int      `json:"start_with_sap"`
			SegmentBase  struct {
				Initialization string `json:"initialization"`
				IndexRange     string `json:"index_range"`
			} `json:"segment_base"`
			Codecid int `json:"codecid"`
		} `json:"video"`
		Audio []struct {
			ID           int      `json:"id"`
			BaseURL      string   `json:"base_url"`
			BackupURL    []string `json:"backup_url"`
			Bandwidth    int      `json:"bandwidth"`
			MimeType     string   `json:"mime_type"`
			Codecs       string   `json:"codecs"`
			Width        int      `json:"width"`
			Height       int      `json:"height"`
			FrameRate    string   `json:"frame_rate"`
			Sar          string   `json:"sar"`
			StartWithSap int      `json:"start_with_sap"`
			SegmentBase  struct {
				Initialization string `json:"initialization"`
				IndexRange     string `json:"index_range"`
			} `json:"segment_base"`
			Codecid int `json:"codecid"`
		} `json:"audio"`
		Dolby struct {
			Type  int `json:"type"`
			Audio any `json:"audio"`
		} `json:"dolby"`
		Flac struct {
			Display bool `json:"display"`
			Audio   struct {
				ID           int    `json:"id"`
				BaseURL      string `json:"base_url"`
				BackupURL    any    `json:"backup_url"`
				Bandwidth    int    `json:"bandwidth"`
				MimeType     string `json:"mime_type"`
				Codecs       string `json:"codecs"`
				Width        int    `json:"width"`
				Height       int    `json:"height"`
				FrameRate    string `json:"frame_rate"`
				Sar          string `json:"sar"`
				StartWithSap int    `json:"start_with_sap"`
				SegmentBase  struct {
					Initialization string `json:"initialization"`
					IndexRange     string `json:"index_range"`
				} `json:"segment_base"`
				Codecid int `json:"codecid"`
			} `json:"audio"`
		} `json:"flac"`
	} `json:"dash"`
	SupportFormats []struct {
		Quality        int      `json:"quality"`
		Format         string   `json:"format"`
		NewDescription string   `json:"new_description"`
		DisplayDesc    string   `json:"display_desc"`
		Superscript    string   `json:"superscript"`
		Codecs         []string `json:"codecs"`
	} `json:"support_formats"`
	HighFormat   any   `json:"high_format"`
	LastPlayTime int   `json:"last_play_time"`
	LastPlayCid  int64 `json:"last_play_cid"`
}

type VideoMetaStruct struct {
	Aid       int64  `json:"aid"`
	Videos    int    `json:"videos"`
	TypeID    int    `json:"type_id"`
	TypeName  string `json:"type_name"`
	Copyright int    `json:"copyright"`
	Pic       string `json:"pic"`
	Title     string `json:"title"`
	Pubdate   int    `json:"pubdate"`
	Ctime     int    `json:"ctime"`
	Desc      string `json:"desc"`
	Duration  int    `json:"duration"`
	MissionID int    `json:"mission_id"`
	Rights    struct {
		Elec      int `json:"elec"`
		Download  int `json:"download"`
		Hd5       int `json:"hd5"`
		NoReprint int `json:"no_reprint"`
		Autoplay  int `json:"autoplay"`
	} `json:"rights"`
	Author struct {
		Mid  int    `json:"mid"`
		Name string `json:"name"`
		Face string `json:"face"`
	} `json:"author"`
	Stat struct {
		Aid     int64 `json:"aid"`
		View    int   `json:"view"`
		Danmaku int   `json:"danmaku"`
		Reply   int   `json:"reply"`
		Fav     int   `json:"fav"`
		Coin    int   `json:"coin"`
		Share   int   `json:"share"`
		Like    int   `json:"like"`
	} `json:"stat"`
	Dynamic   string `json:"dynamic"`
	FirstCid  int64  `json:"first_cid"`
	Dimension struct {
		Width  int `json:"width"`
		Height int `json:"height"`
	} `json:"dimension"`
	ShortLinkV2 string `json:"short_link_v2"`
	FirstFrame  string `json:"first_frame"`
}

type XmlD struct {
	P    string `xml:"p,attr"`
	Text string `xml:",chardata"`
}

type DanmakuXmlstruct struct {
	ChatServer string `xml:"chatserver"`
	ChatID     int64  `xml:"chatid"`
	Mission    int    `xml:"mission"`
	MaxLimit   int    `xml:"maxlimit"`
	State      int    `xml:"state"`
	RealName   int    `xml:"real_name"`
	Source     string `xml:"source"`
	Danmaku    []XmlD `xml:"d"`
}
