# 🚀 5分钟快速上手指南

本指南将帮助您在5分钟内快速上手TTS语音合成应用。

## 📥 第一步：下载应用

### 方式一：下载预编译版本（推荐）

访问 [Releases页面](https://github.com/difyz9/go-tts-app/releases) 下载适合您系统的版本：

```bash
# Linux AMD64
wget https://github.com/difyz9/go-tts-app/releases/latest/download/tts_app_linux_amd64.tar.gz
tar -xzf tts_app_linux_amd64.tar.gz
chmod +x tts_app_linux_amd64

# macOS Intel
wget https://github.com/difyz9/go-tts-app/releases/latest/download/tts_app_darwin_amd64.tar.gz
tar -xzf tts_app_darwin_amd64.tar.gz
chmod +x tts_app_darwin_amd64

# macOS Apple Silicon  
wget https://github.com/difyz9/go-tts-app/releases/latest/download/tts_app_darwin_arm64.tar.gz
tar -xzf tts_app_darwin_arm64.tar.gz
chmod +x tts_app_darwin_arm64

# Windows
# 下载 tts_app_windows_amd64.zip 并解压
```

### 方式二：从源码编译

```bash
# 需要Go 1.23+
git clone https://github.com/difyz9/go-tts-app.git
cd go-tts-app
go build -o tts_app
```

## 📝 第二步：准备文本文件

创建一个包含要转换的文本的文件：

```bash
# 创建测试文件
cat > test.txt << EOF
欢迎使用TTS语音合成应用
这是一个功能强大的文本转语音工具
支持腾讯云TTS和Microsoft Edge TTS
完全免费，开箱即用
EOF
```

## 🎵 第三步：开始转换

### 使用Edge TTS（免费，推荐新手）

```bash
# 最简单的使用方式 - 完全免费！
./tts_app edge -i test.txt

# 指定输出目录
./tts_app edge -i test.txt -o my_output/

# 使用男声
./tts_app edge -i test.txt --voice zh-CN-YunyangNeural

# 调整语速和音量
./tts_app edge -i test.txt --rate +20% --volume +10%
```

### 使用腾讯云TTS（需要API密钥）

```bash
# 1. 复制配置文件
cp config.yaml.example config.yaml

# 2. 编辑配置文件，填入您的腾讯云API密钥
# vim config.yaml 或使用其他编辑器

# 3. 运行转换
./tts_app tts -i test.txt
```

## 📂 第四步：查看结果

转换完成后，您可以在输出目录找到音频文件：

```bash
# 默认输出目录
ls output/
# merged_audio.mp3

# 播放音频（Linux/macOS）
# mpv output/merged_audio.mp3
# 或
# ffplay output/merged_audio.mp3
```

## 🎨 第五步：探索更多功能

### 查看可用的语音

```bash
# 查看所有Edge TTS语音
./tts_app edge --list-all

# 只看中文语音
./tts_app edge --list zh

# 只看英文语音  
./tts_app edge --list en
```

### 自定义语音参数

```bash
# 完整的自定义示例
./tts_app edge -i test.txt \
  --voice zh-CN-XiaoyiNeural \
  --rate +15% \
  --volume +5% \
  --pitch +2Hz \
  -o custom_output/
```

### 合并现有音频文件

```bash
# 如果您已有音频文件需要合并
./tts_app merge --input ./audio_files --output final.mp3
```

## 🔧 配置文件详解

如果您想使用腾讯云TTS或进行更精细的配置：

```yaml
# config.yaml
input_file: "test.txt"

# Edge TTS配置（免费）
edge_tts:
  voice: "zh-CN-XiaoyiNeural"   # 女声
  rate: "+0%"                   # 正常语速
  volume: "+0%"                 # 正常音量

# 腾讯云TTS配置（需要API密钥）
tencent_cloud:
  secret_id: "your_secret_id"
  secret_key: "your_secret_key"
  region: "ap-beijing"

# 输出配置
audio:
  output_dir: "output"
  final_output: "merged_audio.mp3"

# 并发配置
concurrent:
  max_workers: 5
  rate_limit: 20
```

## 💡 使用技巧

### 1. 处理长文本

```bash
# 自动按行分割处理
cat > long_article.txt << EOF
第一段内容...
第二段内容...
第三段内容...
EOF

./tts_app edge -i long_article.txt
```

### 2. 批量处理文件

```bash
# 处理多个文件
for file in *.txt; do
  ./tts_app edge -i "$file" -o "output_${file%.txt}/"
done
```

### 3. 使用不同语言

```bash
# 英文文本
echo "Hello, welcome to TTS application" > english.txt
./tts_app edge -i english.txt --voice en-US-JennyNeural

# 日文文本
echo "こんにちは、TTSアプリケーションへようこそ" > japanese.txt  
./tts_app edge -i japanese.txt --voice ja-JP-NanamiNeural
```

## 🚨 常见问题

### Q: 没有声音输出？
A: 检查：
1. 输入文件是否存在且有内容
2. 网络连接是否正常
3. 输出目录是否有写权限

### Q: Edge TTS失败？
A: Edge TTS需要网络连接，请检查：
1. 网络连接是否正常
2. 是否被防火墙阻止
3. 尝试使用代理（如果在特殊网络环境）

### Q: 想要更多音色？
A: 使用命令查看：
```bash
./tts_app edge --list-all    # 查看所有322个音色
./tts_app edge --list zh     # 查看14个中文音色
```

### Q: 如何处理大文件？
A: 应用会自动：
1. 按行分割文本
2. 并发处理（最多5个并发）
3. 自动合并所有音频

## 🎯 下一步

恭喜！您已经学会了基本用法。接下来可以：

1. 📚 阅读[完整用户手册](../README.md)
2. 🔧 学习[高级配置](advanced-config.md)
3. 🤝 查看[贡献指南](../CONTRIBUTING.md)
4. 💬 加入[社区讨论](https://github.com/difyz9/go-tts-app/discussions)

## 🆘 需要帮助？

- 🐛 [报告Bug](https://github.com/difyz9/go-tts-app/issues/new?template=bug_report.md)
- 💡 [功能请求](https://github.com/difyz9/go-tts-app/issues/new?template=feature_request.md)  
- ❓ [咨询问题](https://github.com/difyz9/go-tts-app/issues/new?template=question.md)
- 💬 [社区讨论](https://github.com/difyz9/go-tts-app/discussions)

---

🎉 **恭喜您完成快速上手！现在您可以享受高质量的语音合成服务了！**
