package service

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// AudioMergeOnlyService 纯音频合并服务
type AudioMergeOnlyService struct{}

// NewAudioMergeOnlyService 创建纯音频合并服务
func NewAudioMergeOnlyService() *AudioMergeOnlyService {
	return &AudioMergeOnlyService{}
}

// MergeAudioFiles 合并音频文件
func (amos *AudioMergeOnlyService) MergeAudioFiles(audioFiles []string, outputPath string) error {
	if len(audioFiles) == 0 {
		return fmt.Errorf("没有音频文件需要合并")
	}

	// 确保输出目录存在
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %v", err)
	}

	// 检查是否所有文件都是相同格式
	if !amos.checkAudioFormatsCompatible(audioFiles) {
		fmt.Println("⚠️  警告: 检测到不同格式的音频文件，合并结果可能不理想")
		fmt.Println("建议使用相同格式的音频文件进行合并")
	}

	// 创建输出文件
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("创建输出文件失败: %v", err)
	}
	defer outputFile.Close()

	// 依次合并音频文件
	for i, audioFile := range audioFiles {
		fmt.Printf("合并文件 %d/%d: %s\n", i+1, len(audioFiles), filepath.Base(audioFile))

		// 检查文件是否存在
		if _, err := os.Stat(audioFile); os.IsNotExist(err) {
			fmt.Printf("⚠️  警告: 文件不存在，跳过: %s\n", audioFile)
			continue
		}

		// 验证音频文件
		if err := amos.validateSingleAudioFile(audioFile); err != nil {
			fmt.Printf("⚠️  警告: 音频文件验证失败，跳过: %s, 错误: %v\n", audioFile, err)
			continue
		}

		// 打开音频文件
		inputFile, err := os.Open(audioFile)
		if err != nil {
			fmt.Printf("⚠️  警告: 打开文件失败，跳过: %s, 错误: %v\n", audioFile, err)
			continue
		}

		// 获取文件大小用于进度显示
		fileInfo, err := inputFile.Stat()
		if err != nil {
			fmt.Printf("⚠️  警告: 获取文件信息失败: %s, 错误: %v\n", audioFile, err)
		} else {
			fmt.Printf("    文件大小: %.2f KB\n", float64(fileInfo.Size())/1024)
		}

		// 复制文件内容
		copied, err := io.Copy(outputFile, inputFile)
		inputFile.Close()

		if err != nil {
			fmt.Printf("⚠️  警告: 复制文件失败，跳过: %s, 错误: %v\n", audioFile, err)
			continue
		}

		fmt.Printf("    已复制: %.2f KB\n", float64(copied)/1024)
	}

	// 获取最终文件大小
	finalInfo, err := outputFile.Stat()
	if err == nil {
		fmt.Printf("\n📊 合并统计:\n")
		fmt.Printf("- 输入文件数: %d\n", len(audioFiles))
		fmt.Printf("- 输出文件: %s\n", outputPath)
		fmt.Printf("- 最终大小: %.2f KB\n", float64(finalInfo.Size())/1024)
	}

	return nil
}

// checkAudioFormatsCompatible 检查音频格式兼容性
func (amos *AudioMergeOnlyService) checkAudioFormatsCompatible(audioFiles []string) bool {
	if len(audioFiles) <= 1 {
		return true
	}

	// 获取第一个文件的扩展名作为基准
	firstExt := strings.ToLower(filepath.Ext(audioFiles[0]))

	// 检查所有文件是否具有相同扩展名
	for _, file := range audioFiles[1:] {
		ext := strings.ToLower(filepath.Ext(file))
		if ext != firstExt {
			return false
		}
	}

	return true
}

// MergeAudioFilesWithFFmpeg 使用FFmpeg合并音频文件（高级版本）
func (amos *AudioMergeOnlyService) MergeAudioFilesWithFFmpeg(audioFiles []string, outputPath string) error {
	// 这个函数预留给未来FFmpeg集成使用
	// 目前使用简单的二进制拼接方式
	fmt.Println("ℹ️  提示: 当前使用简单合并模式")
	fmt.Println("如需高级音频处理，请安装FFmpeg并更新代码")

	return amos.MergeAudioFiles(audioFiles, outputPath)
}

// ValidateAudioFiles 验证音频文件
func (amos *AudioMergeOnlyService) ValidateAudioFiles(audioFiles []string) error {
	for i, file := range audioFiles {
		// 检查文件是否存在
		if _, err := os.Stat(file); os.IsNotExist(err) {
			return fmt.Errorf("文件 %d 不存在: %s", i+1, file)
		}

		// 检查文件是否为音频文件（基于扩展名）
		ext := strings.ToLower(filepath.Ext(file))
		validExtensions := []string{".mp3", ".wav", ".m4a", ".aac", ".flac", ".ogg"}

		isValid := false
		for _, validExt := range validExtensions {
			if ext == validExt {
				isValid = true
				break
			}
		}

		if !isValid {
			return fmt.Errorf("文件 %d 不是支持的音频格式: %s", i+1, file)
		}

		// 检查文件是否可读
		f, err := os.Open(file)
		if err != nil {
			return fmt.Errorf("无法打开文件 %d: %s, 错误: %v", i+1, file, err)
		}
		f.Close()
	}

	return nil
}

// validateSingleAudioFile 验证单个音频文件
func (amos *AudioMergeOnlyService) validateSingleAudioFile(audioPath string) error {
	// 检查文件是否存在
	fileInfo, err := os.Stat(audioPath)
	if err != nil {
		return fmt.Errorf("音频文件不存在: %v", err)
	}

	// 检查文件大小
	const minFileSize = 1024 // 最小1KB
	if fileInfo.Size() < minFileSize {
		return fmt.Errorf("音频文件过小 (%d bytes)，可能为空或损坏", fileInfo.Size())
	}

	// 检查文件是否可读
	file, err := os.Open(audioPath)
	if err != nil {
		return fmt.Errorf("无法打开音频文件: %v", err)
	}
	defer file.Close()

	// 读取文件头部进行基本格式验证
	buffer := make([]byte, 12)
	n, err := file.Read(buffer)
	if err != nil || n < 4 {
		return fmt.Errorf("无法读取音频文件头部")
	}

	// 获取文件扩展名
	ext := strings.ToLower(filepath.Ext(audioPath))
	
	// 根据扩展名验证文件头部
	switch ext {
	case ".mp3":
		if n >= 3 && (string(buffer[:3]) == "ID3" || 
			(buffer[0] == 0xFF && (buffer[1]&0xF0) == 0xF0)) {
			return nil
		}
		return fmt.Errorf("文件头部不匹配MP3格式")
	case ".wav":
		if n >= 12 && string(buffer[:4]) == "RIFF" && string(buffer[8:12]) == "WAVE" {
			return nil
		}
		return fmt.Errorf("文件头部不匹配WAV格式")
	case ".m4a", ".aac":
		// M4A/AAC文件通常以ftyp开头（在前8字节后）
		if n >= 8 {
			return nil // 简化验证，只检查大小
		}
		return fmt.Errorf("文件头部读取不足")
	case ".flac":
		if n >= 4 && string(buffer[:4]) == "fLaC" {
			return nil
		}
		return fmt.Errorf("文件头部不匹配FLAC格式")
	case ".ogg":
		if n >= 4 && string(buffer[:4]) == "OggS" {
			return nil
		}
		return fmt.Errorf("文件头部不匹配OGG格式")
	default:
		// 对于未知格式，只检查是否为空文件
		return nil
	}
}
