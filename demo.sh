#!/bin/bash

echo "=== TTS语音合成应用演示 ==="
echo

# 检查配置文件
if [ ! -f "config.yaml" ]; then
    echo "错误: 未找到配置文件 config.yaml"
    echo "请先配置腾讯云密钥信息"
    exit 1
fi

# 检查历史文件
if [ ! -f "example_input.txt" ]; then
    echo "错误: 未找到输入文件 example_input.txt"
    exit 1
fi

echo "1. 检查配置文件..."
echo "✓ 配置文件: config.yaml"
echo "✓ 输入文件: example_input.txt"
echo

echo "2. 当前配置信息:"
echo "   - 输入文件: example_input.txt"
echo "   - 输出目录: ./output"
echo "   - 临时目录: ./temp"
echo "   - 最终文件: merged_audio.mp3"
echo

echo "3. 输入文件内容预览:"
head -5 example_input.txt
echo "   ..."
echo "   (共 $(wc -l < example_input.txt) 行)"
echo

echo "4. 开始TTS转换..."
echo "注意: 此演示需要有效的腾讯云密钥"
echo "如果没有配置密钥，程序会显示配置错误"
echo

# 运行TTS程序
./tts_app tts --config config.yaml

if [ $? -eq 0 ]; then
    echo
    echo "=== 处理完成 ==="
    echo "检查输出文件:"
    ls -la output/
    echo
    echo "检查临时文件:"
    ls -la temp/ | head -10
else
    echo
    echo "=== 处理失败 ==="
    echo "请检查:"
    echo "1. 腾讯云密钥配置是否正确"
    echo "2. 网络连接是否正常"
    echo "3. 文件权限是否足够"
fi
