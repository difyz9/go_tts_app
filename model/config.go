package model

// Config 总配置结构
type Config struct {
	TencentCloud TencentCloudConfig `yaml:"tencent_cloud"`
	TTS          TTSConfig          `yaml:"tts"`
	EdgeTTS      EdgeTTSConfig      `yaml:"edge_tts"`
	Audio        AudioConfig        `yaml:"audio"`
	Concurrent   ConcurrentConfig   `yaml:"concurrent"`
	InputFile    string             `yaml:"input_file"`
}

// TencentCloudConfig 腾讯云配置
type TencentCloudConfig struct {
	SecretID  string `yaml:"secret_id"`
	SecretKey string `yaml:"secret_key"`
	Region    string `yaml:"region"`
}

// TTSConfig TTS音频参数配置
type TTSConfig struct {
	VoiceType       int64   `yaml:"voice_type"`
	Volume          int64   `yaml:"volume"`
	Speed           float64 `yaml:"speed"`
	PrimaryLanguage int64   `yaml:"primary_language"`
	SampleRate      int64   `yaml:"sample_rate"`
	Codec           string  `yaml:"codec"`
}

// EdgeTTSConfig Edge TTS配置
type EdgeTTSConfig struct {
	Voice  string `yaml:"voice"`   // 语音名称，如 zh-CN-XiaoyiNeural
	Rate   string `yaml:"rate"`    // 语速，如 +10%, +0%, -10%
	Volume string `yaml:"volume"`  // 音量，如 +10%, +0%, -10%
	Pitch  string `yaml:"pitch"`   // 音调，如 +10Hz, +0Hz, -10Hz
}

// AudioConfig 音频合并配置
type AudioConfig struct {
	OutputDir       string  `yaml:"output_dir"`
	TempDir         string  `yaml:"temp_dir"`
	FinalOutput     string  `yaml:"final_output"`
	SilenceDuration float64 `yaml:"silence_duration"`
}

// ConcurrentConfig 并发配置
type ConcurrentConfig struct {
	MaxWorkers int `yaml:"max_workers"`
	RateLimit  int `yaml:"rate_limit"`
	BatchSize  int `yaml:"batch_size"`
}
