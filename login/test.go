package login

import (
	"github.com/rs/zerolog/log"

	"github.com/XiaoMiku01/bilibili-archiver/internal"
)

func RunTest(config internal.Config) {
	internal.BApi.SetCookieFile(config.User)

	buser, err := internal.BApi.GetUserInfo()
	if err != nil {
		log.Fatal().Err(err).Msg("获取用户信息失败")
	}
	log.Info().Msgf("用户: %s [UID: %d] 登录成功", buser.Uname, buser.Mid)

	// 测试通知
	if config.Notification != "" {
		err = internal.SendNotification(config.Notification, "B站留档助手 测试通知", config.NotificationProxy)
		if err != nil {
			log.Error().Err(err).Msg("发送通知失败")
		} else {
			log.Info().Msg("发送通知成功")
		}
	}
}
