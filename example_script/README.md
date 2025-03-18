## 自定义脚本说明  

程序执行自定义脚本时，首先会在标准输入中写入完成留档的路径目录（而非视频文件的路径），脚本中读取一行输入后再执行其他命令（例如遍历目录）

- [Xml 弹幕转 ass](xml2ass.sh)  
需要下载 [DanmakuFactory](https://github.com/hihkm/DanmakuFactory) 作为转换工具，放在工作目录下，赋予执行权限，注意下载对应系统的二进制文件

其他功能可自行实现，如果本地有 python 环境，同样支持 python 脚本，nodejs 同理