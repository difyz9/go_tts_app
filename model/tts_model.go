package model

// TTS合成请求
type TTSRequest struct {
	Text            string  `json:"text" binding:"required"`
	VoiceType       int64   `json:"voiceType,omitempty"`
	Volume          int64   `json:"volume,omitempty"`
	Speed           float64 `json:"speed,omitempty"` // 修改为float64类型
	PrimaryLanguage int64   `json:"primaryLanguage,omitempty"`
	SampleRate      int64   `json:"sampleRate,omitempty"`
	Codec           string  `json:"codec,omitempty"`
}

// TTS任务响应
type TTSResponse struct {
	Success bool   `json:"success"`
	TaskID  string `json:"taskId,omitempty"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// TTS任务状态查询响应
type TTSStatusResponse struct {
	Success   bool   `json:"success"`
	Status    int64  `json:"status,omitempty"`
	StatusStr string `json:"statusStr,omitempty"`
	AudioURL  string `json:"audioUrl,omitempty"`
	ErrorMsg  string `json:"errorMsg,omitempty"`
	Error     string `json:"error,omitempty"`
}

// 健康检查响应
type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Version   string `json:"version"`
}
