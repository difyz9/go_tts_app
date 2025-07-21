package main

import (
	"fmt"
	"tts_app/service"
)

func main() {
	processor := service.NewTextProcessor()
	
	testCases := []string{
		"func main() {",
		"    fmt.Println(\"Hello\")",
		"return",
		"这是正常文本",
		"package main",
		"import \"fmt\"",
		"if (condition) {",
		"} else {",
		"for i := 0; i < 10; i++ {",
	}
	
	fmt.Println("=== 代码过滤测试 ===")
	for i, text := range testCases {
		isValid := processor.IsValidTextForTTS(text)
		fmt.Printf("%d. \"%s\" -> %v\n", i+1, text, isValid)
	}
}
