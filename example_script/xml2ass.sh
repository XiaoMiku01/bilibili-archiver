#!/bin/bash

# 从标准输入读取路径
read input_path

# 检查路径是否存在
if [ ! -d "$input_path" ]; then
    echo "错误：指定的路径不存在或不是一个目录"
    exit 1
fi

# 查找所有xml文件并处理
find "$input_path" -type f -name "*.xml" | while read -r xml_file; do
    # 生成对应的ass文件路径（替换.xml为.ass并删除_danmaku）
    ass_file="${xml_file%.xml}"
    ass_file="${ass_file/_danmaku/}.ass"
    
    # 执行转换命令
    ./DanmakuFactory -i "$xml_file" -o "$ass_file" > /dev/null 2>&1
    
    # 检查命令是否成功执行
    if [ $? -eq 0 ]; then
        echo "$ass_file 已完成转换"
    else
        echo "转换失败: $xml_file"
    fi
done