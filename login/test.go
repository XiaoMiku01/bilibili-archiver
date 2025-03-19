package login

import (
	"os"
	"os/exec"

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
	tinfo, _ := internal.BApi.CheckToken()
	exTime := tinfo.ExpiresIn / 86400
	log.Info().Msgf("Cookie 有效期: %d 天", exTime)
	CheckFFmpeg()
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

func CheckFFmpeg() {
	// 首先在 PATH 环境变量中查找 ffmpeg
	ffmpegPath, err := exec.LookPath("ffmpeg")
	if err == nil {
		log.Info().Msgf("发现 ffmpeg: %s", ffmpegPath)
		return
	}

	// 如果 PATH 中没有找到，检查当前目录
	_, err = os.Stat("./ffmpeg")
	if err == nil {
		log.Info().Msg("在当前目录中找到 ffmpeg")
		return
	}

	// 都没找到，输出错误
	log.Fatal().Msg("环境检查失败: 未找到 ffmpeg，请确保已安装 ffmpeg 并添加到环境变量")
}
