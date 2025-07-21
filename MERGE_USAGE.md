# 音频合并功能使用指南

## merge命令功能

merge命令可以将指定目录下的音频文件按照不同的排序方式合并成一个音频文件。

## 基本语法

```bash
./tts_app merge --input <输入目录> --output <输出文件> [选项]
```

## 参数说明

| 参数 | 短参数 | 必需 | 说明 | 默认值 |
|------|--------|------|------|--------|
| `--input` | `-i` | ✅ | 输入目录路径 | - |
| `--output` | `-o` | ✅ | 输出文件路径 | - |
| `--sort` | - | ❌ | 排序方式 | `name` |
| `--format` | - | ❌ | 音频格式 | `mp3` |

## 排序方式

### 1. 按文件名排序 (`name`)
```bash
./tts_app merge -i ./temp -o merged_by_name.mp3 --sort name
```
- 按字母顺序排序文件名
- 适用于有序命名的文件（如 audio_001.mp3, audio_002.mp3）

### 2. 按修改时间排序 (`time`)
```bash
./tts_app merge -i ./temp -o merged_by_time.mp3 --sort time
```
- 按文件修改时间从早到晚排序
- 适用于按生成时间顺序合并

### 3. 按文件大小排序 (`size`)
```bash
./tts_app merge -i ./temp -o merged_by_size.mp3 --sort size
```
- 按文件大小从小到大排序
- 可以用于特殊的音频处理需求

## 支持的音频格式

输入格式：
- `.mp3` - MP3音频文件
- `.wav` - WAV音频文件  
- `.m4a` - M4A音频文件
- `.aac` - AAC音频文件
- `.flac` - FLAC音频文件
- `.ogg` - OGG音频文件

输出格式：
- 由输出文件的扩展名决定
- 建议使用与输入文件相同的格式

## 使用示例

### 示例1：基本合并
```bash
# 合并temp目录下的所有音频文件
./tts_app merge --input ./temp --output ./output/merged.mp3
```

### 示例2：完整工作流程
```bash
# 1. 首先生成TTS音频
./tts_app tts --concurrent

# 2. 查看生成的文件
ls -la temp/

# 3. 合并音频文件
./tts_app merge -i ./temp -o ./output/final_audio.mp3

# 4. 验证结果
ls -la output/
```

## 输出信息说明

程序会显示以下信息：

1. **配置信息**：显示输入目录、输出文件、排序方式等
2. **文件列表**：按合并顺序显示所有找到的音频文件
3. **合并进度**：显示每个文件的处理进度和大小
4. **统计信息**：显示输入文件数量、输出文件和最终大小
