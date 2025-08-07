# 统一语音合成接口使用指南

## 概述

本项目实现了统一的语音合成接口，支持多个语音合成提供商，包括腾讯云TTS和Microsoft Edge TTS。通过统一的接口设计，可以轻松切换不同的语音合成服务，并且所有服务都具有统一的文本处理、速率限制和并发控制功能。

## 核心特性

### 1. 统一接口设计
- **TTSProvider接口**: 定义了所有语音合成提供商必须实现的方法
- **UnifiedTTSService**: 提供统一的服务入口，封装了通用逻辑

### 2. 统一的文本处理
- 自动清理和标准化文本格式
- 智能文本分割，避免超长文本问题
- Markdown文档专业解析
- 无效文本过滤

### 3. 统一的速率限制
- 可配置的并发控制
- 每个提供商有推荐的速率限制
- 智能重试机制

### 4. 并发处理
- 多worker并发生成音频
- 任务队列管理
- 结果收集和排序

## 接口定义

```go
type TTSProvider interface {
    // 生成音频，返回音频文件路径
    GenerateAudio(ctx context.Context, text string, index int) (string, error)
    
    // 获取提供商名称
    GetProviderName() string
    
    // 验证配置是否正确
    ValidateConfig() error
    
    // 获取单次请求最大文本长度
    GetMaxTextLength() int
    
    // 获取推荐的速率限制（每秒请求数）
    GetRecommendedRateLimit() int
}
```

## 支持的提供商

### 1. 腾讯云TTS (TencentTTSProvider)
- **提供商名称**: "TencentCloud"
- **最大文本长度**: 150字符
- **推荐速率限制**: 5请求/秒
- **配置要求**: SecretID, SecretKey, Region

### 2. Microsoft Edge TTS (EdgeTTSProvider)
- **提供商名称**: "EdgeTTS"
- **最大文本长度**: 1000字符
- **推荐速率限制**: 10请求/秒
- **配置要求**: 无特殊要求

## 使用方法

### 1. 基本使用

```go
// 创建腾讯云TTS服务
tencentService, err := CreateUnifiedTTSService("tencent", config)
if err != nil {
    return err
}

// 处理Markdown文件
err = tencentService.ProcessMarkdownFile("input.md", "output/")
if err != nil {
    return err
}
```

### 2. 切换提供商

```go
// 切换到Edge TTS
edgeService, err := CreateUnifiedTTSService("edge", config)
if err != nil {
    return err
}

// 处理普通文本文件
err = edgeService.ProcessInputFile("input.txt", "output/")
if err != nil {
    return err
}
```

### 3. 工厂模式创建

```go
factory := &TTSProviderFactory{}

// 根据配置创建提供商
provider, err := factory.CreateProvider("tencent", config)
if err != nil {
    return err
}

// 创建统一服务
service := NewUnifiedTTSService(provider, config)
```

## 配置文件

### config.yaml 示例

```yaml
# 腾讯云配置
tencent_cloud:
  secret_id: "your_secret_id"
  secret_key: "your_secret_key"
  region: "ap-beijing"

# TTS参数配置
tts:
  voice_type: 101008
  volume: 5
  speed: 1.0
  primary_language: 1
  sample_rate: 16000
  codec: "mp3"

# Edge TTS配置
edge_tts:
  voice: "zh-CN-XiaoyiNeural"
  rate: "+0%"
  volume: "+0%"
  pitch: "+0Hz"

# 音频输出配置
audio:
  output_dir: "output"
  temp_dir: "temp"
  final_output: "merged_audio.mp3"

# 并发配置
concurrent:
  max_workers: 5
  rate_limit: 5
  batch_size: 10

input_file: "input.txt"
```

## 扩展新的提供商

要添加新的语音合成提供商，只需要实现 `TTSProvider` 接口：

```go
type CustomTTSProvider struct {
    config *model.Config
}

func (ctp *CustomTTSProvider) GenerateAudio(ctx context.Context, text string, index int) (string, error) {
    // 实现音频生成逻辑
    return audioPath, nil
}

func (ctp *CustomTTSProvider) GetProviderName() string {
    return "Custom"
}

func (ctp *CustomTTSProvider) ValidateConfig() error {
    // 验证配置
    return nil
}

func (ctp *CustomTTSProvider) GetMaxTextLength() int {
    return 500 // 自定义最大长度
}

func (ctp *CustomTTSProvider) GetRecommendedRateLimit() int {
    return 8 // 自定义速率限制
}
```

然后在工厂中注册新的提供商：

```go
func (factory *TTSProviderFactory) CreateProvider(providerType string, config *model.Config) (TTSProvider, error) {
    switch providerType {
    case "custom":
        return NewCustomTTSProvider(config), nil
    // ... 其他提供商
    }
}
```

## 优势

1. **统一的用户体验**: 不管使用哪个提供商，调用方式都是相同的
2. **配置驱动**: 可以通过配置文件轻松切换提供商
3. **自动优化**: 每个提供商都有针对性的优化参数
4. **易于扩展**: 添加新提供商只需实现接口
5. **健壮性**: 统一的错误处理、重试机制和验证逻辑
6. **性能优化**: 智能并发控制和速率限制

## 最佳实践

1. **根据需求选择提供商**:
   - 腾讯云TTS: 商业项目，需要高质量语音
   - Edge TTS: 个人项目，免费使用

2. **合理配置并发参数**:
   - 腾讯云TTS: 建议较低的并发数和速率限制
   - Edge TTS: 可以使用较高的并发数

3. **文本预处理**:
   - 系统会自动处理文本，但预先清理可以提高效果
   - 对于Markdown文档，建议使用专用的处理方法

4. **错误处理**:
   - 系统内置重试机制，但仍需要处理最终失败的情况
   - 关注速率限制和配额问题
