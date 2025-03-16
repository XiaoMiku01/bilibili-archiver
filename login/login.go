package login

import (
	"fmt"
	"time"

	"github.com/XiaoMiku01/bilibili-archiver/internal"
	"github.com/rs/zerolog/log"
)

func Run() {
	qr, err := internal.BApi.GetQRCode()
	if err != nil {
		log.Error().Msg("获取二维码失败")
		return
	}
	log.Info().Str("QRCodeUrl", qr.Url).Msg("获取二维码成功")
	qrImg := internal.NewQR(qr.Url)
	if err := qrImg.Print(); err != nil {
		log.Error().Err(err).Msg("打印二维码失败")
		return
	}
	var cookieInfo internal.CookieInfoStruct
	for {
		time.Sleep(3 * time.Second)
		cookieInfo, err = internal.BApi.VerifyQrCode(qr)
		if err != nil {
			switch err := err.(type) {
			case *internal.BiliErr:
				if err.Code == 86039 {
					break
				} else if err.Code == 86038 {
					log.Error().Err(err).Msg("二维码已失效")
					break
				} else {
					log.Error().Err(err).Msg("验证二维码失败")
				}
			default:
				log.Error().Err(err).Msg("验证二维码失败")
			}
		}
		if cookieInfo.Mid != 0 {
			log.Info().Msg("验证二维码成功")
			break
		}
	}
	cfName := fmt.Sprintf("%d_cookie.json", cookieInfo.Mid)
	cookieInfo.SaveToFile(cfName)
	internal.RefreshToken(cfName)
	if uf, err := internal.CheckCookieFile(cfName); err {
		log.Info().Msgf("%s UID: %v 登录成功 cookie 文件保存在 %s", uf.Uname, uf.Mid, cfName)
	}
}
