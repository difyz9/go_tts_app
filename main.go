/*
TTS语音合成应用
支持腾讯云TTS和Microsoft Edge TTS的高性能文本转语音工具

Copyright © 2025 TTS App Contributors
*/
package main

import (
	"tts_app/cmd"
)

// 版本信息，在编译时通过ldflags注入
var (
	version   = "dev"
	buildTime = "unknown"
	gitCommit = "unknown"
)

func main() {
	// 设置版本信息到cmd包
	cmd.SetVersionInfo(version, buildTime, gitCommit)
	
	// 执行根命令
	cmd.Execute()
}
