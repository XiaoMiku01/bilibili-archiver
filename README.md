# B站留档工具  

自动同步B站收藏夹投稿、弹幕、元数据至本地  

## 实现功能

- [x] 扫码登录, 自动保活账号
- [x] 同步下载收藏夹投稿、弹幕
- [x] 收藏夹关键词过滤
- [x] 定时更新数据
- [ ] 多渠道发送通知
- [ ] 自定义留档后脚本
- [ ] 支持 Docker 部署

## 配置文件示例  

```yaml
user: "./cookie.json"  # 用户的 cookie 文件
save_path: "./videos"  # 存储目录

keywords:  # 收藏夹的关键词，如果为空则全部同步
  - "留档"
  - "备份"

scan_interval: 5  # 扫描收藏夹间隔 (分钟)
update_interval: 10  # 更新元数据时间 (分钟)
update_dl : 7 # 投稿发布后多久停止更新元数据 (天)

incremental: true  # 是否开启增量同步（只同步启动后增加的内容）
danmaku: false  # 是否同时下载弹幕

notification: "telegram://token@telegram?chats=@channel-1,chat-id-1"  
notification_proxy : "http://127.0.0.1:7890" # 通知使用的代理
# 支持的通知种类和示例见: https://containrrr.dev/shoutrrr/v0.8/services/overview/
# 触发通知事件：（通知测试，新投稿留档成功、失败，已留档投稿失效）
custom_script: "bash run.sh"  # 自定义存档成功后的脚本

```
