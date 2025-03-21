# B站留档助手  

[![Release Version](https://img.shields.io/github/v/release/XiaoMiku01/bilibili-archiver?style=flat-square)](https://github.com/XiaoMiku01/bilibili-archiver/releases/latest)
[![License](https://img.shields.io/github/license/XiaoMiku01/bilibili-archiver?style=flat-square)](https://github.com/XiaoMiku01/bilibili-archiver/blob/main/LICENSE)
[![Go Version](https://img.shields.io/github/go-mod/go-version/XiaoMiku01/bilibili-archiver?style=flat-square)](https://go.dev/)
[![Go Report Card](https://goreportcard.com/badge/github.com/XiaoMiku01/bilibili-archiver)](https://goreportcard.com/report/github.com/XiaoMiku01/bilibili-archiver)

自动同步B站收藏夹投稿、弹幕、元数据至本地  

<a href="https://asciinema.org/a/708893" target="_blank"><img src="./demo.gif" /></a>  

## 目录

- [实现功能](#实现功能)
- [TODO](#todo)
- [使用方法](#使用方法)
  - [安装](#安装)
    - [方法一: 直接下载可执行文件](#方法一-直接下载可执行文件)
    - [方法二: 从源码编译](#方法二-从源码编译)
  - [快速开始](#快速开始)
  - [命令说明](#命令说明)
    - [全局参数](#全局参数)
    - [可用命令](#可用命令)
  - [Docker 部署](#docker-部署)
- [注意事项](#注意事项)
- [配置文件示例](#配置文件示例)
- [第三方库和参考项目](#第三方库和参考项目)

## 实现功能

- [x] 扫码登录, 自动保活账号
- [x] 同步下载收藏夹投稿、弹幕
- [x] 收藏夹关键词过滤
- [x] 定时更新数据
- [x] 多渠道发送通知
- [x] 自定义留档后、更新元数据脚本
- [x] 收藏夹投稿失效通知
- [x] 下载视频时规避 PCDN
- [x] 支持 Docker 部署

## TODO

- [ ] 对指定 UP 主持续监控
- [ ] ~WebUI 播放本地视频~ 建议使用 [alist](https://github.com/AlistGo/alist) 挂载本地目录(支持本地弹幕和在线弹幕)

## 使用方法

### 安装

#### 方法一: 直接下载可执行文件

从 [Release页面](https://github.com/XiaoMiku01/bilibili-archiver/releases) 下载对应系统的可执行文件。

#### 方法二: 从源码编译

```bash
git clone https://github.com/XiaoMiku01/bilibili-archiver.git
cd bilibili-archiver
go build .
```


### 快速开始

1. 首先登录获取 cookie

```bash
./bilibili-archiver login
```

2. 按照提示使用手机扫码完成登录

3. 创建配置文件 `config.yaml`(参考配置文件示例)

4. 测试配置是否正确:

```bash
./bilibili-archiver test
```

5. 启动程序:

```bash
./bilibili-archiver start
```

### 命令说明

```bash
bilibili-archiver [<flags>] <command> [<args>...]
```

#### 全局参数

- `-h, --help`: 显示帮助信息
- `-v, --debug`: 开启调试日志输出
- `-c, --config="./config.yaml"`: 指定配置文件路径，默认为当前目录下的 config.yaml

#### 可用命令

- `login`: 扫码登录B站获取 <uid>_cookie.json
- `test`: 测试登录状态和通知渠道配置
- `refresh [<flags>]`: 更新 cookie.json 保持登录状态
  - `-u, --cookie=COOKIE`: 指定要刷新的 cookie 文件
- `start`: 开始运行程序，按照配置自动同步收藏夹内容

### Docker 部署

1. 拉取镜像

```bash
docker pull ghcr.io/xiaomiku01/bilibili-archiver:latest
# 或者本地构建
docker build -t bilibili-archiver .
```

2. 创建容器

```bash
# 创建工作目录
mkdir bilibili-archiver 
cd bilibili-archiver

# 登录
docker run --rm \
  -v $(pwd):/data \
  ghcr.io/xiaomiku01/bilibili-archiver login

# 启动
docker run -d \
  --name bilibili-archiver \
  -v $(pwd):/data \
  ghcr.io/xiaomiku01/bilibili-archiver start
```

## 注意事项

**FFmpeg 依赖**：本项目依赖 FFmpeg 合并音视频，请确保系统已安装 FFmpeg。  
Docker 镜像中已经包含 FFmpeg 


## 配置文件示例  

```yaml
user: cookie.json  # 用户的 cookie 文件
save_path: ./videos  # 存储目录

# 保存路径模板
# {{ uname }} - 用户名
# {{ fav_name }} - 收藏夹名
# {{ date }} - 收藏日期
# {{ video_title }} - 投稿标题
# {{ bv }} - 投稿BV号
# {{ upper_name }} - up主名
# {{ pn }} - 投稿分p序号
# / 为路径分隔符
# 例如: {{ uname }}/{{ fav_name }}/{{ video_title }}.{{ upper_name }}/{{ bv }}-P{{ pn }}[{{ video_quality }}]
path_template: "{{ uname }}/{{ fav_name }}/{{ date }}-{{ video_title }}.{{ upper_name }}/{{ bv }}-P{{ pn }}"

keywords:  # 收藏夹的关键词，如果为空则全部同步
  - 留档
  - 备份

scan_interval: 10  # 扫描收藏夹间隔 (分钟)
update_interval: 30  # 更新元数据时间 (分钟)
update_dl : 7 # 投稿发布后多久停止更新元数据 (天)

incremental: true  # 是否开启增量同步（只同步启动后增加的内容）, 如果关闭第一次同步会同步所有投稿
danmaku: true  # 是否同时下载弹幕

# 支持的通知种类和示例见: https://containrrr.dev/shoutrrr/v0.8/services/overview/
notification: telegram://token@telegram?chats=@channel-1,chat-id-1
notification_proxy : "" # 通知使用的代理 支持 socks5:// 和 http://

custom_script: ""  # 自定义存档成功后的脚本 如xml转ass脚本  bash example_script/xml2ass.sh 
run_after_update: ""  # 更新元数据后运行的脚本 可以和上面的脚本一样 用于将新增的弹幕转为ass

disable_pcdn: false  # 禁用PCDN下载视频 PCDN下载可能会导致视频花屏
```

[示例自定义脚本](./example_script/)

### 第三方库和参考项目  

[developer](./developer.md)
