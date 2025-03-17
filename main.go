package main

import (
	"os"

	"github.com/alecthomas/kingpin/v2"
	"github.com/rs/zerolog/log"

	"github.com/XiaoMiku01/bilibili-archiver/archiver"
	"github.com/XiaoMiku01/bilibili-archiver/internal"
	"github.com/XiaoMiku01/bilibili-archiver/login"
)

var (
	Version = "v0.0.1"

	app = kingpin.New("bilibili-archiver", "B站留档助手")
	// 全局标志
	debug  = app.Flag("debug", "开启debug日志打印").Short('v').Bool()
	config = app.Flag("config", "配置文件").Short('c').Default("./config.yaml").String()

	// login 命令
	loginCmd = app.Command("login", "扫码登录B站获取 <uid>_cookie.json")
	testCmd  = app.Command("test", "测试登录，通知渠道（若配置）")

	// refresh 命令
	refreshCmd = app.Command("refresh", "更新 cookie.json")

	cookieFile = refreshCmd.Flag("cookie", "cookie文件").Short('u').String()

	// start 命令
	startCmd = app.Command("start", "开始运行程序")

	// test 命令
)

func main() {
	app.HelpFlag.Short('h').Help("显示帮助信息")
	command := kingpin.MustParse(app.Parse(os.Args[1:]))

	// 初始化日志记录器
	internal.InitLogger(*debug)

	log.Debug().Bool("Debug:", *debug).Msg("开启debug日志")
	log.Info().Str("Version:", Version).Msg("B站留档助手")
	log.Info().Str("Config:", *config).Msg("配置文件")

	switch command {
	case loginCmd.FullCommand():
		log.Info().Msg("开始登录")
		login.Run()

	case refreshCmd.FullCommand():
		log.Info().Msg("开始刷新 Cookie")
		internal.RefreshToken(*cookieFile)

	case startCmd.FullCommand():
		config, err := internal.LoadConfig(*config)
		if err != nil {
			log.Fatal().Err(err).Msg("加载配置文件失败")
		}
		log.Info().Msg("开始运行")
		archiver := archiver.NewArchiverUser(*config)
		err = archiver.Init()
		if err != nil {
			log.Fatal().Err(err).Msg("初始化用户失败")
		}
		// TODO: 实现启动逻辑
		archiver.Run()

	case testCmd.FullCommand():
		log.Info().Msg("测试配置")
		config, err := internal.LoadConfig(*config)
		if err != nil {
			log.Fatal().Err(err).Msg("加载配置文件失败")
		}
		login.RunTest(*config)
		// TODO: 实现测试逻辑
	}
}
