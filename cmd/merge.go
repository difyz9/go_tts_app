/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"github.com/difyz9/markdown2tts/service"

	"github.com/spf13/cobra"
)

var (
	inputDir    string
	outputFile  string
	audioFormat string
)

// mergeCmd represents the merge command
var mergeCmd = &cobra.Command{
	Use:   "merge",
	Short: "合并指定目录下的音频文件",
	Long: `将指定目录下的音频文件按照文件名中的数字顺序合并成一个音频文件。

自动提取文件名中的数字进行排序，例如：
- audio_001.mp3, audio_002.mp3, audio_010.mp3
- sound1.wav, sound2.wav, sound10.wav

支持的音频格式：mp3, wav, m4a等

示例:
  github.com/difyz9/markdown2tts merge --input ./temp --output merged.mp3
  github.com/difyz9/markdown2tts merge --input ./audio_files --output final.wav`,
	Run: func(cmd *cobra.Command, args []string) {
		err := runMerge()
		if err != nil {
			fmt.Printf("错误: %v\n", err)
			os.Exit(1)
		}
	},
}

func runMerge() error {
	// 验证输入参数
	if inputDir == "" {
		return fmt.Errorf("请指定输入目录 --input")
	}
	if outputFile == "" {
		return fmt.Errorf("请指定输出文件 --output")
	}

	// 检查输入目录是否存在
	if _, err := os.Stat(inputDir); os.IsNotExist(err) {
		return fmt.Errorf("输入目录不存在: %s", inputDir)
	}

	fmt.Printf("合并配置:\n")
	fmt.Printf("- 输入目录: %s\n", inputDir)
	fmt.Printf("- 输出文件: %s\n", outputFile)
	fmt.Printf("- 排序方式: 按文件名数字顺序\n")
	fmt.Printf("- 音频格式: %s\n", audioFormat)
	fmt.Println()

	// 创建音频合并服务
	mergeService := service.NewAudioMergeOnlyService()

	// 扫描并收集音频文件
	audioFiles, err := scanAudioFiles(inputDir)
	if err != nil {
		return fmt.Errorf("扫描音频文件失败: %v", err)
	}

	if len(audioFiles) == 0 {
		return fmt.Errorf("在目录 %s 中没有找到音频文件", inputDir)
	}

	fmt.Printf("找到 %d 个音频文件\n", len(audioFiles))

	// 按文件名数字顺序排序
	sortAudioFilesByNumber(audioFiles)

	// 显示文件列表
	fmt.Println("\n音频文件列表（按数字顺序）:")
	for i, file := range audioFiles {
		fmt.Printf("%d. %s (数字: %d)\n", i+1, filepath.Base(file.Path), file.Number)
	}
	fmt.Println()

	// 提取文件路径
	filePaths := make([]string, len(audioFiles))
	for i, file := range audioFiles {
		filePaths[i] = file.Path
	}

	// 合并音频文件
	fmt.Println("开始合并音频文件...")
	err = mergeService.MergeAudioFiles(filePaths, outputFile)
	if err != nil {
		return fmt.Errorf("合并音频文件失败: %v", err)
	}

	fmt.Printf("✅ 音频合并完成: %s\n", outputFile)
	return nil
}

// AudioFileInfo 音频文件信息
type AudioFileInfo struct {
	Path   string
	Name   string
	Number int // 从文件名提取的数字，用于排序
}

// scanAudioFiles 扫描目录中的音频文件
func scanAudioFiles(dir string) ([]AudioFileInfo, error) {
	var audioFiles []AudioFileInfo

	// 支持的音频格式
	audioExtensions := map[string]bool{
		".mp3":  true,
		".wav":  true,
		".m4a":  true,
		".aac":  true,
		".flac": true,
		".ogg":  true,
	}

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录
		if info.IsDir() {
			return nil
		}

		// 检查文件扩展名
		ext := strings.ToLower(filepath.Ext(path))
		if audioExtensions[ext] {
			// 提取文件名中的数字
			number := extractNumberFromFilename(info.Name())

			audioFiles = append(audioFiles, AudioFileInfo{
				Path:   path,
				Name:   info.Name(),
				Number: number,
			})
		}

		return nil
	})

	return audioFiles, err
}

// extractNumberFromFilename 从文件名中提取数字
func extractNumberFromFilename(filename string) int {
	// 移除文件扩展名
	nameWithoutExt := strings.TrimSuffix(filename, filepath.Ext(filename))

	// 使用正则表达式提取所有数字
	re := regexp.MustCompile(`\d+`)
	matches := re.FindAllString(nameWithoutExt, -1)

	if len(matches) == 0 {
		// 如果没有找到数字，返回一个很大的数，让它排在最后
		return 999999
	}

	// 优先提取以下划线分隔的数字（如audio_001.mp3中的001）
	// 或者取最长的数字序列
	var bestMatch string
	maxLength := 0

	for _, match := range matches {
		if len(match) > maxLength {
			maxLength = len(match)
			bestMatch = match
		}
	}

	// 如果没有找到最佳匹配，取最后一个数字
	if bestMatch == "" {
		bestMatch = matches[len(matches)-1]
	}

	number, err := strconv.Atoi(bestMatch)
	if err != nil {
		return 999999 // 转换失败时也排在最后
	}

	return number
}

// sortAudioFilesByNumber 按文件名中的数字排序，数字相同时按文件名排序
func sortAudioFilesByNumber(audioFiles []AudioFileInfo) {
	sort.Slice(audioFiles, func(i, j int) bool {
		// 首先按数字排序
		if audioFiles[i].Number != audioFiles[j].Number {
			return audioFiles[i].Number < audioFiles[j].Number
		}
		// 数字相同时按文件名排序
		return audioFiles[i].Name < audioFiles[j].Name
	})
}

func init() {
	rootCmd.AddCommand(mergeCmd)

	// 添加命令行参数
	mergeCmd.Flags().StringVarP(&inputDir, "input", "i", "", "输入目录路径（必需）")
	mergeCmd.Flags().StringVarP(&outputFile, "output", "o", "", "输出文件路径（必需）")
	mergeCmd.Flags().StringVar(&audioFormat, "format", "mp3", "音频格式 (mp3, wav, m4a等)")

	// 标记必需参数
	mergeCmd.MarkFlagRequired("input")
	mergeCmd.MarkFlagRequired("output")
}
