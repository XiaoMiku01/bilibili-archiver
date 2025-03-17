# B站留档工具  

自动同步B站收藏夹投稿、弹幕、元数据至本地  

## 实现功能

- [x] 扫码登录, 自动保活账号
- [x] 同步下载收藏夹投稿、弹幕
- [x] 收藏夹关键词过滤
- [x] 定时更新数据
- [x] 多渠道发送通知
- [x] 自定义留档后脚本
- [x] 支持 Docker 部署

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



```
