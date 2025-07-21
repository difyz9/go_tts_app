# 🎵 TTS语音合成应用

[![Go Version](https://img.shields.io/badge/Go-1.23+-blue.svg)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Build Status](https://github.com/difyz9/markdown2tts/workflows/Build/badge.svg)](https://github.com/difyz9/markdown2tts/actions)
[![Release](https://img.shields.io/github/release/difyz9/go-tts-app.svg)](https://github.com/difyz9/markdown2tts/releases)

一个功能完整、高性能的文本转语音(TTS)应用程序，支持**双引擎**、**并发处理**、**智能过滤**等特色功能。

## ✨ 核心特色

### 🎯 **双引擎支持**
- **腾讯云TTS** - 企业级，高质量音色，支持精细参数调节
- **Microsoft Edge TTS** - 完全免费，无需API密钥，开箱即用

### 🚀 **高性能特性**
- **智能并发处理** - 最高20倍速度提升
- **智能限流控制** - 自动控制API请求频率
- **实时进度显示** - 详细的处理状态和统计信息
- **批量音频合并** - 自动合并为完整音频文件

### 🔧 **智能特性**
- **智能文本过滤** - 自动跳过空行、标记行、短文本 
- **特殊字符处理** - 智能处理Markdown格式、转义字符、中英文混合等
- **灵活配置管理** - 支持配置文件和命令行参数
- **多格式支持** - 支持MP3、WAV等多种音频格式
- **跨平台支持** - Windows、macOS、Linux全平台兼容

## 🚀 快速开始

### 方式一：一键初始化（推荐新用户）
```bash
# 下载并解压最新版本
wget https://github.com/difyz9/markdown2tts/releases/latest/download/tts_app_linux_amd64.tar.gz
tar -xzf tts_app_linux_amd64.tar.gz

# 初始化配置文件和示例文件
./tts_app init

# 立即开始转换（完全免费）
./tts_app edge -i input.txt
```

### 方式二：Edge TTS（免费，无需配置）
```bash
# 创建测试文件
echo "欢迎使用TTS应用，这是一个完全免费的语音合成工具" > test.txt

# 立即开始转换（配置文件会自动创建）
./tts_app edge -i test.txt

# 智能Markdown模式（推荐用于.md文件）
./tts_app edge -i document.md --smart-markdown -o output

# 传统模式（用于纯文本文件）
./tts_app edge -i document.txt -o output

```

### 方式三：腾讯云TTS（企业用户）
```bash
# 初始化配置
./tts_app init

# 编辑 config.yaml，填入腾讯云密钥
nano config.yaml

# 使用腾讯云TTS
./tts_app tts -i input.txt
```

## 📋 命令详解

### Edge TTS 命令（免费推荐）
```bash
# 基本使用
./tts_app edge                              # 使用默认配置
./tts_app edge -i input.txt                 # 指定输入文件
./tts_app edge -i input.txt -o output/      # 指定输出目录

# 查看可用语音
./tts_app edge --list-all                   # 显示所有语音（322个）
./tts_app edge --list zh                    # 显示中文语音（14个）
./tts_app edge --list en                    # 显示英文语音（47个）

# 自定义语音参数
./tts_app edge --voice zh-CN-YunyangNeural          # 使用男声
./tts_app edge --voice zh-CN-XiaoyiNeural           # 使用女声
./tts_app edge --rate +20% --volume +10%            # 调整语速和音量
./tts_app edge --voice zh-CN-YunyangNeural --rate +15% --volume +5% --pitch +5Hz  # 完整自定义
```

### 腾讯云TTS 命令
```bash
# 基本使用
./tts_app tts                               # 使用默认配置
./tts_app tts --config custom.yaml          # 使用自定义配置
./tts_app tts -i input.txt -o output/       # 指定输入输出

# 并发处理（默认开启）
./tts_app tts --concurrent                  # 明确启用并发模式
```

### 音频合并命令
```bash
# 合并音频文件
./tts_app merge --input ./temp --output merged.mp3

```

## ⚙️ 配置说明

### 基础配置文件 (config.yaml)

```yaml
# 输入文件配置
input_file: "example_input.txt"      # 默认输入文件路径

# 腾讯云TTS配置（企业用户）
tencent_cloud:
  secret_id: "your_secret_id"
  secret_key: "your_secret_key"
  region: "ap-beijing"

# TTS音频参数
tts:
  voice_type: 101008      # 音色ID：101008-智琪(女声), 101007-智慧(女声), 101003-智云(男声)
  volume: 5               # 音量：0-10
  speed: 1.0              # 语速：0.6-1.5
  primary_language: 1     # 主语言：1-中文，2-英文
  sample_rate: 16000      # 采样率
  codec: "mp3"            # 编码格式

# Edge TTS配置（免费用户）
edge_tts:
  voice: "zh-CN-XiaoyiNeural"   # 语音名称
  rate: "+0%"                   # 语速调节：-50% 到 +100%
  volume: "+0%"                 # 音量调节：-50% 到 +100%
  pitch: "+0Hz"                 # 音调调节：-50Hz 到 +50Hz

# 音频处理配置
audio:
  output_dir: "output"
  temp_dir: "temp"
  final_output: "merged_audio.mp3"
  silence_duration: 0.5

# 并发处理配置
concurrent:
  max_workers: 5          # 最大并发数
  rate_limit: 20          # 每秒请求限制
  batch_size: 10          # 批处理大小

```


## 🔧 智能初始化

### 自动配置创建
应用首次运行时会自动检测并创建必需的配置文件：

```bash
# 首次运行任何命令时都会自动初始化
./tts_app edge -i your_text.txt   # 自动创建 config.yaml 和 input.txt
./tts_app tts -i your_text.txt    # 同样会自动初始化
```

### 手动初始化
如果需要手动初始化或重新初始化：

```bash
# 基本初始化
./tts_app init

# 自定义文件名
./tts_app init --config my_config.yaml --input my_input.txt

# 强制覆盖已存在的文件
./tts_app init --force
```

### 初始化内容
- **config.yaml** - 完整的配置文件模板
- **input.txt** - 包含示例内容的输入文件
- **自动提示** - 详细的使用指南和下一步操作


## 🎯 使用场景

### 个人用户
- **学习材料制作** - 将文章、笔记转换为音频
- **有声读物制作** - 制作个人有声书
- **语言学习** - 制作外语学习材料

### 企业用户
- **客服语音** - 生成客服自动语音
- **产品介绍** - 制作产品语音介绍
- **培训材料** - 制作员工培训音频

### 开发者
- **应用集成** - 为应用添加语音功能
- **自动化流程** - 批量处理文本内容
- **多媒体制作** - 制作播客、视频配音

## 🔥 性能对比

| 特性 | 传统方式 | 本应用 |
|------|----------|--------|
| 处理速度 | 顺序处理，慢 | 并发处理，快20倍 |
| 费用成本 | 昂贵API费用 | Edge TTS完全免费 |
| 配置复杂度 | 复杂配置 | 一键启动 |
| 文本过滤 | 手动处理 | 智能过滤 |
| 进度显示 | 无反馈 | 实时进度 |
| 错误处理 | 容易失败 | 智能重试 |

## 📊 功能对比

| 功能 | 腾讯云TTS | Edge TTS |
|------|-----------|----------|
| 费用 | 按量收费 | 🆓 完全免费 |
| 配置 | 需要API密钥 | 🚀 无需配置 |
| 音色数量 | 丰富(50+) | 中等(322) |
| 质量 | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| 并发支持 | ✅ | ✅ |
| 智能过滤 | ✅ | ✅ |
| 实时进度 | ✅ | ✅ |


## 📁 项目结构

```
go-tts-app/
├── main.go                 # 程序入口
├── config.yaml            # 配置文件
├── example_input.txt         # 示例输入文件
├── cmd/                   # 命令行接口
│   ├── root.go           # 根命令
│   ├── tts.go            # 腾讯云TTS命令
│   ├── edge.go           # Edge TTS命令
│   └── merge.go          # 音频合并命令
├── model/                 # 数据模型
│   ├── config.go         # 配置结构
│   └── tts_model.go      # TTS模型
├── service/               # 核心服务
│   ├── tts_service.go           # 腾讯云TTS服务
│   ├── edge_tts_service.go      # Edge TTS服务
│   ├── concurrent_audio_service.go  # 并发处理服务
│   ├── audio_service.go         # 音频处理服务
│   └── audio_merge_only_service.go  # 音频合并服务
├── output/                # 输出目录
├── temp/                  # 临时文件目录
└── docs/                  # 文档目录

```


## 🛠️ 开发构建

### 环境要求
- Go 1.23+
- 网络连接

### 本地开发
```bash
# 克隆项目
git clone https://github.com/difyz9/markdown2tts.git
cd go-tts-app

# 安装依赖
go mod download

# 构建项目
go build -o tts_app

# 运行测试
go test ./...

# 本地运行
./tts_app edge --help
```

### 交叉编译
```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o tts_app_linux_amd64
GOOS=linux GOARCH=arm64 go build -o tts_app_linux_arm64

# Windows
GOOS=windows GOARCH=amd64 go build -o tts_app_windows_amd64.exe

# macOS
GOOS=darwin GOARCH=amd64 go build -o tts_app_darwin_amd64
GOOS=darwin GOARCH=arm64 go build -o tts_app_darwin_arm64
```

## 🎵 音色示例

### Edge TTS 推荐中文音色
- **zh-CN-XiaoyiNeural** - 晓伊，女声，温和自然
- **zh-CN-YunyangNeural** - 云扬，男声，成熟稳重
- **zh-CN-XiaochenNeural** - 晓辰，女声，活泼可爱
- **zh-CN-YunxiNeural** - 云希，男声，年轻阳光

### 腾讯云TTS 音色ID
- **101008** - 智琪，女声，知性优雅
- **101007** - 智慧，女声，亲和温柔
- **101003** - 智云，男声，磁性深沉
## 📚 使用教程

### 新手教程
1. [5分钟快速上手](docs/quick-start.md)
2. [Edge TTS完整指南](docs/edge-tts-guide.md)
3. [常见问题解答](docs/faq.md)

### 进阶教程
1. [腾讯云TTS配置](docs/tencent-setup.md)
2. [并发优化技巧](docs/performance-tips.md)
3. [批量处理最佳实践](docs/batch-processing.md)

## 🤝 贡献指南

我们欢迎所有形式的贡献！

### 如何贡献
1. Fork 本项目
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 创建 Pull Request

### 贡献类型
- 🐛 Bug 修复
- ✨ 新功能开发
- 📚 文档改进
- 🎨 UI/UX 改进
- ⚡ 性能优化
- 🧪 测试覆盖

## 📄 开源协议

本项目采用 [MIT License](LICENSE) 开源协议。

## 🌟 致谢

- [腾讯云TTS](https://cloud.tencent.com/product/tts) - 提供企业级TTS服务
- [Microsoft Edge TTS](https://azure.microsoft.com/en-us/services/cognitive-services/text-to-speech/) - 提供免费TTS服务
- [Cobra](https://github.com/spf13/cobra) - 强大的CLI框架
- [edge-tts-go](https://github.com/difyz9/edge-tts-go) - Edge TTS Go实现

## 📞 支持与反馈

- 🐛 [报告Bug](https://github.com/difyz9/markdown2tts/issues/new?template=bug_report.md)
- 💡 [功能建议](https://github.com/difyz9/markdown2tts/issues/new?template=feature_request.md)
- 📖 [查看文档](https://github.com/difyz9/markdown2tts/wiki)
- 💬 [讨论交流](https://github.com/difyz9/markdown2tts/discussions)

## 📈 项目状态

- ✅ 基础TTS功能
- ✅ 双引擎支持
- ✅ 并发处理
- ✅ 智能过滤
- ✅ 音频合并
- ✅ 跨平台支持
- 🚧 Web界面 (开发中)
- 🚧 API服务 (规划中)
- 🚧 Docker支持 (规划中)

---

⭐ 如果这个项目对您有帮助，请给我们一个星标！

🔔 点击 Watch 获取项目更新通知

1. **配置文件错误**
   - 检查config.yaml格式是否正确
   - 确认腾讯云密钥配置正确

2. **网络错误**
   - 检查网络连接
   - 确认腾讯云服务地域设置

3. **文件权限错误**
   - 确保程序有读取ai_history.txt的权限
   - 确保程序有创建输出目录和文件的权限

4. **TTS任务失败**
   - 检查文本内容是否包含特殊字符
   - 确认TTS参数设置是否正确

### 调试模式

程序会输出详细的处理日志，包括：
- 读取的文本行数
- 每行文本的处理状态
- TTS任务创建和状态查询
- 音频下载和合并进度

## 开发说明

项目使用Go语言开发，主要依赖：

- **cobra**: 命令行框架
- **yaml.v3**: YAML配置文件解析
- **tencentcloud-sdk-go**: 腾讯云Go SDK

### 项目结构

- `cmd/`: 命令行接口
- `model/`: 数据模型定义
- `service/`: 业务逻辑服务
  - `tts_service.go`: 腾讯云TTS API封装
  - `audio_service.go`: 音频处理和合并服务
  - `config.yaml`: 配置文件服务

### 扩展功能

可以考虑的扩展功能：
- 支持更多TTS服务商
- 使用FFmpeg进行高级音频处理
- 添加语音情感控制
- 支持SSML标记语言
- 添加音频效果处理
