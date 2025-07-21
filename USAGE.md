# TTS应用使用指南

## 快速开始

### 腾讯云TTS（需要API密钥）

#### 最简单的使用方式
```bash
# 默认使用ai_history.txt作为输入，输出到./output目录
./tts_app tts
```

#### 指定输入文件
```bash
# 指定要转换的文本文件
./tts_app tts -i my_text.txt
```

#### 指定输出目录
```bash
# 指定输出目录
./tts_app tts -i my_text.txt -o /path/to/output
```

### Edge TTS（免费，无需API密钥）

#### 最简单的使用方式
```bash
# 使用免费的Microsoft Edge TTS服务
./tts_app edge
```

#### 指定输入文件
```bash
# 指定要转换的文本文件
./tts_app edge -i my_text.txt
```

#### 指定输出目录
```bash
# 指定输出目录
./tts_app edge -i my_text.txt -o /path/to/output
```

#### 查看可用语音
```bash
# 列出所有可用语音
./tts_app edge --list

# 只显示中文语音
./tts_app edge --list zh

# 只显示英文语音
./tts_app edge --list en
```

#### 自定义语音参数
```bash
# 使用指定语音
./tts_app edge -i input.txt --voice zh-CN-YunyangNeural

# 调整语速和音量
./tts_app edge -i input.txt --rate +20% --volume +10%

# 完整自定义
./tts_app edge -i input.txt --voice zh-CN-YunyangNeural --rate +20% --volume +10% --pitch +5Hz
```

### 使用自定义配置文件
```bash
# 腾讯云TTS使用自定义配置文件
./tts_app tts --config custom_config.yaml

# Edge TTS使用自定义配置文件
./tts_app edge --config custom_config.yaml
```

## 功能特点

### 腾讯云TTS
- **高质量音色**: 支持多种精品音色
- **自定义参数**: 可调节语速、音量、音调
- **需要配置**: 需要腾讯云API密钥

### Edge TTS
- **完全免费**: 无需任何API密钥或费用
- **高质量合成**: 使用Microsoft Edge的在线TTS服务
- **多语言支持**: 支持中文、英文等多种语言
- **即开即用**: 无需配置，直接使用

### 通用特性
- **默认并发模式**: 自动启用并发处理，提高转换效率
- **自动配置加载**: 默认自动查找当前目录的config.yaml文件
- **简化参数**: 只需要指定最关键的输入和输出参数
- **智能路径处理**: 自动处理相对和绝对路径
- **目录自动创建**: 输出目录不存在时自动创建

## 配置文件

### 腾讯云TTS配置
应用会自动查找当前目录下的`config.yaml`文件，包含：
- 腾讯云API凭证（必需）
- TTS参数（音色、语速、音量）
- 并发配置
- 音频格式设置

### Edge TTS配置
Edge TTS使用相同的配置文件，但不需要API凭证：
- 并发配置
- 音频格式设置
- 输入输出路径配置

## 输入文件格式

文本文件应该是：
- UTF-8编码
- 每行一个句子或段落
- 空行会被自动跳过（包括只包含空格的行）
- 支持任意长度的文本行

### 空行处理
应用会智能跳过以下类型的行：
- 完全空白的行
- 只包含空格、制表符等空白字符的行
- 以 `###` 开头的标题行（Markdown标题）
- 以 `**` 开头的加粗标记行
- 以 `-----` 开头的分隔线
- 在处理过程中显示跳过的空行和标记行数量统计

### 支持的标记格式
```
### 这是标题                    # 跳过
** 这是加粗标记                # 跳过  
----- 这是分隔线               # 跳过
正常文本中包含###符号           # 处理
正常文本中包含**符号            # 处理
正常文本中包含-----符号         # 处理
   ### 前面有空格的标题         # 跳过
```

## 输出结果

- 临时音频文件存储在`temp/`目录
- 最终合并的音频文件输出到指定目录
- 默认输出文件名：`merged_audio.mp3`

## 示例

```bash
# 腾讯云TTS基本使用
./tts_app tts

# Edge TTS基本使用（推荐）
./tts_app edge

# 转换特定文件到特定目录
./tts_app edge -i article.txt -o podcast/

# 使用腾讯云TTS的自定义配置
./tts_app tts --config production.yaml -i script.txt -o final/

# 使用Edge TTS（免费）
./tts_app edge -i chinese_text.txt -o voice_output/
```

## 推荐使用方式

**新用户推荐**: 使用 `./tts_app edge` 命令，完全免费，无需配置。

**企业用户**: 使用 `./tts_app tts` 命令，支持更多音色选择和参数调节。

## 处理结果示例

```
读取到 15 行文本，开始并发生成音频...
跳过 11 个空行/标记行，实际处理 4 行有效文本
启动 4 个worker开始处理...
✓ 任务 1 完成: temp/audio_001.mp3
✓ 任务 4 完成: temp/audio_004.mp3
✓ 任务 8 完成: temp/audio_008.mp3
✓ 任务 14 完成: temp/audio_014.mp3

处理完成: 成功 4, 失败 0
```
