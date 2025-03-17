package internal

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	User              string   `yaml:"user"`               // cookie文件路径
	SavePath          string   `yaml:"save_path"`          // 投稿存储目录
	PathTemplate      string   `yaml:"path_template"`      // 存储路径模板
	Keywords          []string `yaml:"keywords"`           // 收藏夹关键词过滤
	ScanInterval      int      `yaml:"scan_interval"`      // 扫描收藏夹间隔(分钟)
	UpdateInterval    int      `yaml:"update_interval"`    // 更新元数据间隔(分钟)
	UpdateDL          int      `yaml:"update_dl"`          // 停止更新元数据的天数
	Incremental       bool     `yaml:"incremental"`        // 是否开启增量同步
	Danmaku           bool     `yaml:"danmaku"`            // 是否下载弹幕
	Notification      string   `yaml:"notification"`       // 通知配置
	NotificationProxy string   `yaml:"notification_proxy"` // 通知代理
	CustomScript      string   `yaml:"custom_script"`      // 自定义脚本
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	config := &Config{}
	err = yaml.Unmarshal(data, config)
	if err != nil {
		return nil, err
	}

	// 设置默认值
	if config.SavePath == "" {
		config.SavePath = "./videos"
	}
	if config.PathTemplate == "" {
		config.PathTemplate = "{{ uname }}/{{ fav_name }}/{{ date }}-{{ video_title }}.{{ upper_name }}/{{ bv }}-P{{ pn }}"
	}
	if config.ScanInterval <= 0 {
		config.ScanInterval = 10 // 默认10分钟
	}
	if config.UpdateInterval <= 0 {
		config.UpdateInterval = 30 // 默认10分钟
	}
	if config.UpdateDL <= 0 {
		config.UpdateDL = 7 // 默认7天
	}

	// 统一打印配置信息
	fmt.Println("当前配置信息:")
	fmt.Println("- cookie文件路径:", config.User)
	fmt.Println("- 投稿存储目录:", config.SavePath)
	fmt.Println("- 存储路径模板:", config.PathTemplate)
	fmt.Println("- 收藏夹关键词过滤:", config.Keywords)
	fmt.Println("- 扫描收藏夹间隔:", config.ScanInterval, "分钟")
	fmt.Println("- 更新元数据间隔:", config.UpdateInterval, "分钟")
	fmt.Println("- 停止更新元数据的天数:", config.UpdateDL, "天")
	fmt.Println("- 是否开启增量同步:", config.Incremental)
	fmt.Println("- 是否下载弹幕:", config.Danmaku)
	fmt.Println("- 通知配置:", config.Notification)
	fmt.Println("- 通知代理:", config.NotificationProxy)
	fmt.Println("- 自定义脚本:", config.CustomScript)

	return config, nil
}
