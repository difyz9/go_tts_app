package service

import (
	"regexp"
	"strings"
	"unicode"
)

// TextProcessor 文本预处理器
type TextProcessor struct {
	// 配置选项
	preserveMarkdown     bool
	normalizeWhitespace  bool
	handleSpecialSymbols bool
	markdownProcessor    *MarkdownProcessor // 新增：专业的Markdown处理器
}

// NewTextProcessor 创建新的文本处理器
func NewTextProcessor() *TextProcessor {
	return &TextProcessor{
		preserveMarkdown:     true,
		normalizeWhitespace:  true,
		handleSpecialSymbols: true,
		markdownProcessor:    NewMarkdownProcessor(), // 初始化Markdown处理器
	}
}

// ProcessText 处理文本，优化TTS语音合成效果
func (tp *TextProcessor) ProcessText(text string) string {
	if text == "" {
		return text
	}

	// 1. 移除Markdown中不需要语音合成的内容（代码块、表格、图片、链接等）
	text = tp.removeNonSpeechElements(text)

	// 2. 处理转义字符（需要在Markdown处理之前）
	text = tp.processEscapeCharacters(text)

	// 3. 处理Markdown格式字符
	if tp.preserveMarkdown {
		text = tp.processMarkdownFormatting(text)
	}

	// 4. 处理特殊符号
	if tp.handleSpecialSymbols {
		text = tp.processSpecialSymbols(text)
	}

	// 5. 规范化空白字符
	if tp.normalizeWhitespace {
		text = tp.normalizeWhitespaceText(text)
	}

	// 6. 处理中英文混合文本
	text = tp.processMixedLanguageText(text)

	// 7. 处理各种类型的括号
	text = tp.processBrackets(text)

	return text
}

// ProcessMarkdownDocument 使用专业Markdown解析器处理整个文档
func (tp *TextProcessor) ProcessMarkdownDocument(markdown string) []string {
	// 使用专业的Markdown处理器提取纯文本
	extractedText := tp.markdownProcessor.ExtractTextForTTS(markdown)

	// 分割成适合TTS的句子
	sentences := tp.markdownProcessor.SplitIntoSentences(extractedText)

	// 对每个句子进行进一步的文本处理
	var processedSentences []string
	for _, sentence := range sentences {
		if sentence == "" {
			continue
		}

		// 使用现有的文本处理逻辑
		processed := tp.ProcessText(sentence)
		if processed != "" && tp.IsValidTextForTTS(processed) {
			processedSentences = append(processedSentences, processed)
		}
	}

	return processedSentences
}

// removeNonSpeechElements 移除Markdown中不需要语音合成的元素
func (tp *TextProcessor) removeNonSpeechElements(text string) string {
	// 1. 移除代码块（``` 或 ~~~ 包围的内容）
	text = tp.removeCodeBlocks(text)

	// 2. 移除表格
	text = tp.removeTables(text)

	// 3. 移除图片
	text = tp.removeImages(text)

	// 4. 处理链接（保留文本，移除URL）
	text = tp.processLinks(text)

	// 5. 移除HTML标签
	text = tp.removeHTMLTags(text)

	// 6. 移除其他Markdown元素
	text = tp.removeOtherMarkdownElements(text)

	return text
}

// removeCodeBlocks 移除代码块
func (tp *TextProcessor) removeCodeBlocks(text string) string {
	// 移除三个反引号包围的代码块（支持语言标识符）
	// 修改正则表达式以更好地匹配代码块边界
	codeBlockRegex := regexp.MustCompile("(?s)```[a-zA-Z0-9]*\\s*\\n.*?\\n```\\s*")
	text = codeBlockRegex.ReplaceAllString(text, "\n")

	// 移除三个波浪号包围的代码块
	tildeCodeBlockRegex := regexp.MustCompile("(?s)~~~[a-zA-Z0-9]*\\s*\\n.*?\\n~~~\\s*")
	text = tildeCodeBlockRegex.ReplaceAllString(text, "\n")

	// 移除单行代码块（行首4个空格缩进）
	indentedCodeRegex := regexp.MustCompile("(?m)^    .*$")
	text = indentedCodeRegex.ReplaceAllString(text, "")

	return text
}

// removeTables 移除Markdown表格
func (tp *TextProcessor) removeTables(text string) string {
	// 移除Markdown表格（包含 | 分隔符的行）
	lines := strings.Split(text, "\n")
	var filteredLines []string

	inTable := false
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// 检查是否是表格行（包含管道符 |）
		if strings.Contains(trimmedLine, "|") && tp.isTableRow(trimmedLine) {
			inTable = true
			continue // 跳过表格行
		}

		// 检查表格分隔符行（如 |---|---|）
		if tp.isTableSeparator(trimmedLine) {
			inTable = true
			continue
		}

		// 如果前一行是表格，当前行不是表格，则表格结束
		if inTable && !strings.Contains(trimmedLine, "|") {
			inTable = false
		}

		// 不在表格中的行保留
		if !inTable {
			filteredLines = append(filteredLines, line)
		}
	}

	return strings.Join(filteredLines, "\n")
}

// removeImages 移除图片
func (tp *TextProcessor) removeImages(text string) string {
	// 移除Markdown图片格式 ![alt](url) 或 ![alt](url "title")
	imageRegex := regexp.MustCompile(`!\[([^\]]*)\]\([^)]+\)`)
	text = imageRegex.ReplaceAllString(text, "")

	// 移除HTML img标签
	htmlImageRegex := regexp.MustCompile(`(?i)<img[^>]*>`)
	text = htmlImageRegex.ReplaceAllString(text, "")

	return text
}

// processLinks 处理链接（保留链接文本，移除URL）
func (tp *TextProcessor) processLinks(text string) string {
	// 处理Markdown链接格式 [text](url)，保留text部分
	linkRegex := regexp.MustCompile(`\[([^\]]+)\]\([^)]+\)`)
	text = linkRegex.ReplaceAllString(text, "$1")

	// 移除纯URL（http://、https://、ftp://、www.）
	urlRegex := regexp.MustCompile(`https?://[^\s]+|ftp://[^\s]+|www\.[^\s]+`)
	text = urlRegex.ReplaceAllString(text, "")

	// 移除邮箱地址
	emailRegex := regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
	text = emailRegex.ReplaceAllString(text, "")

	return text
}

// removeHTMLTags 移除HTML标签
func (tp *TextProcessor) removeHTMLTags(text string) string {
	// 移除HTML标签但保留内容
	htmlTagRegex := regexp.MustCompile(`<[^>]*>`)
	text = htmlTagRegex.ReplaceAllString(text, "")

	// 移除HTML实体
	htmlEntityRegex := regexp.MustCompile(`&[a-zA-Z0-9#]+;`)
	text = htmlEntityRegex.ReplaceAllStringFunc(text, func(entity string) string {
		// 转换常见HTML实体
		entities := map[string]string{
			"&amp;":   "&",
			"&lt;":    "<",
			"&gt;":    ">",
			"&quot;":  "\"",
			"&apos;":  "'",
			"&nbsp;":  " ",
			"&copy;":  "版权",
			"&reg;":   "注册商标",
			"&trade;": "商标",
		}
		if replacement, exists := entities[entity]; exists {
			return replacement
		}
		return "" // 其他实体直接移除
	})

	return text
}

// removeOtherMarkdownElements 移除其他Markdown元素
func (tp *TextProcessor) removeOtherMarkdownElements(text string) string {
	// 移除水平分割线
	hrRegex := regexp.MustCompile(`(?m)^[-*_]{3,}\s*$`)
	text = hrRegex.ReplaceAllString(text, "")

	// 移除引用块标记（保留内容）
	blockquoteRegex := regexp.MustCompile(`(?m)^>\s*`)
	text = blockquoteRegex.ReplaceAllString(text, "")

	// 移除任务列表标记
	taskListRegex := regexp.MustCompile(`(?m)^[-*+]\s*\[[x\s]\]\s*`)
	text = taskListRegex.ReplaceAllString(text, "")

	// 移除普通列表标记（保留内容）
	listRegex := regexp.MustCompile(`(?m)^[-*+]\s+`)
	text = listRegex.ReplaceAllString(text, "")

	// 移除有序列表标记（保留内容）
	orderedListRegex := regexp.MustCompile(`(?m)^\d+\.\s+`)
	text = orderedListRegex.ReplaceAllString(text, "")

	return text
}

// isTableRow 检查是否是表格行
func (tp *TextProcessor) isTableRow(line string) bool {
	// 简单检查：包含管道符且不是代码
	if !strings.Contains(line, "|") {
		return false
	}

	// 排除代码块中的管道符
	if strings.HasPrefix(strings.TrimSpace(line), "```") || strings.HasPrefix(strings.TrimSpace(line), "~~~") {
		return false
	}

	// 检查是否有足够的表格特征（至少2个管道符）
	pipeCount := strings.Count(line, "|")
	return pipeCount >= 2
}

// isTableSeparator 检查是否是表格分隔符行（如 |---|---|）
func (tp *TextProcessor) isTableSeparator(line string) bool {
	// 表格分隔符特征：包含 |、- 和可能的 :
	if !strings.Contains(line, "|") || !strings.Contains(line, "-") {
		return false
	}

	// 移除空格后检查
	cleaned := strings.ReplaceAll(line, " ", "")

	// 检查是否符合表格分隔符模式
	separatorRegex := regexp.MustCompile(`^\|?(:?-+:?\|)+:?-+:?\|?$`)
	return separatorRegex.MatchString(cleaned)
}

// processMarkdownFormatting 处理Markdown格式字符
func (tp *TextProcessor) processMarkdownFormatting(text string) string {
	// 处理加粗标记 **text**
	// 保留内容，移除markdown标记，但保留用于TTS的适当停顿
	boldRegex := regexp.MustCompile(`\*\*([^*]+)\*\*`)
	text = boldRegex.ReplaceAllString(text, "$1")

	// 处理斜体标记 *text*
	italicRegex := regexp.MustCompile(`\*([^*]+)\*`)
	text = italicRegex.ReplaceAllString(text, "$1")

	// 处理代码块标记 `code`
	codeRegex := regexp.MustCompile("`([^`]+)`")
	text = codeRegex.ReplaceAllString(text, "$1")

	// 处理标题标记 ### title
	headerRegex := regexp.MustCompile(`^#+\s*(.+)$`)
	text = headerRegex.ReplaceAllString(text, "$1")

	// 处理链接标记 [text](url)
	linkRegex := regexp.MustCompile(`\[([^\]]+)\]\([^)]+\)`)
	text = linkRegex.ReplaceAllString(text, "$1")

	return text
}

// processEscapeCharacters 处理转义字符
func (tp *TextProcessor) processEscapeCharacters(text string) string {
	// 处理常见的转义序列
	replacements := map[string]string{
		`\*`: "*",  // 转义的星号
		`\\`: "\\", // 转义的反斜杠
		`\n`: " ",  // 换行符转为空格
		`\t`: " ",  // 制表符转为空格
		`\r`: "",   // 回车符删除
		`\"`: "\"", // 转义的双引号
		`\'`: "'",  // 转义的单引号
		`\&`: "&",  // 转义的&符号
		`\#`: "#",  // 转义的#符号
		`\%`: "%",  // 转义的%符号
		`\$`: "$",  // 转义的$符号
		`\@`: "@",  // 转义的@符号
		`\!`: "!",  // 转义的!符号
		`\?`: "?",  // 转义的?符号
		`\+`: "+",  // 转义的+符号
		`\=`: "=",  // 转义的=符号
		`\-`: "-",  // 转义的-符号
		`\_`: "_",  // 转义的下划线
		`\^`: "^",  // 转义的^符号
		`\~`: "~",  // 转义的~符号
		`\|`: "|",  // 转义的|符号
		`\>`: ">",  // 转义的>符号
		`\<`: "<",  // 转义的<符号
		`\{`: "{",  // 转义的{符号
		`\}`: "}",  // 转义的}符号
		`\[`: "[",  // 转义的[符号
		`\]`: "]",  // 转义的]符号
		`\(`: "(",  // 转义的(符号
		`\)`: ")",  // 转义的)符号
	}

	for escaped, unescaped := range replacements {
		text = strings.ReplaceAll(text, escaped, unescaped)
	}

	return text
}

// processSpecialSymbols 处理特殊符号
func (tp *TextProcessor) processSpecialSymbols(text string) string {
	// 为一些特殊符号添加适当的语音停顿或读法
	// 只有当符号独立存在且不在常见上下文中时才替换
	symbolReplacements := map[string]string{
		"@": "at",
		"#": "井号",
		"$": "美元",
		"%": "百分号",
		"^": "",
		"&": "",
		"*": "",
		"+": "加",
		"=": "等于",
		"|": "",
		"~": "",
		"`": "",
		"<": "小于",
		">": "大于",
		"[": "左方括号",
		"]": "右方括号",
		"{": "左大括号",
		"}": "右大括号",
	}

	// 只替换独立的符号，避免破坏有意义的文本
	for symbol, replacement := range symbolReplacements {
		// 更精确的匹配：符号前后必须是空格、标点或字符串边界
		// 但要避免替换有意义的组合，如邮箱、网址、价格等
		pattern := `(\s|^)` + regexp.QuoteMeta(symbol) + `(\s|$)`
		regex := regexp.MustCompile(pattern)
		text = regex.ReplaceAllStringFunc(text, func(match string) string {
			// 检查是否在特殊上下文中（如邮箱、网址、价格等）
			if tp.isInSpecialContext(text, symbol, match) {
				return match // 保持原样
			}
			return strings.Replace(match, symbol, replacement, 1)
		})
	}

	return text
}

// isInSpecialContext 检查符号是否在特殊上下文中（如邮箱、网址等）
func (tp *TextProcessor) isInSpecialContext(text, symbol, match string) bool {
	// 检查常见的特殊上下文模式
	specialPatterns := []string{
		`\w+@\w+\.\w+`,               // 邮箱地址
		`https?://[^\s]+`,            // 网址
		`\$\d+`,                      // 价格（美元）
		`\d+%`,                       // 百分比
		`\d+\.\d+`,                   // 小数
		`#[a-zA-Z_]\w*`,              // 编程中的标识符
		`\*+[^*]*\*+`,                // 被星号包围的文本
		`\+\d+(-\d+)*`,               // 电话号码
		`[a-zA-Z0-9]+\.[a-zA-Z0-9]+`, // 域名或文件扩展名
	}

	for _, pattern := range specialPatterns {
		if matched, _ := regexp.MatchString(pattern, text); matched {
			return true
		}
	}

	return false
}

// normalizeWhitespaceText 规范化空白字符
func (tp *TextProcessor) normalizeWhitespaceText(text string) string {
	// 替换多个连续空格为单个空格
	spaceRegex := regexp.MustCompile(`\s+`)
	text = spaceRegex.ReplaceAllString(text, " ")

	// 移除行首行尾空格
	text = strings.TrimSpace(text)

	return text
}

// processMixedLanguageText 处理中英文混合文本
func (tp *TextProcessor) processMixedLanguageText(text string) string {
	var result strings.Builder
	runes := []rune(text)

	for i, r := range runes {
		// 检查当前字符类型
		isChinese := tp.isChinese(r)
		isEnglish := tp.isEnglish(r)

		// 在中英文之间添加适当的停顿
		if i > 0 {
			prevR := runes[i-1]
			prevIsChinese := tp.isChinese(prevR)
			prevIsEnglish := tp.isEnglish(prevR)

			// 中文后跟英文，或英文后跟中文时，确保有空格分隔
			if (prevIsChinese && isEnglish) || (prevIsEnglish && isChinese) {
				if result.Len() > 0 && !unicode.IsSpace(prevR) {
					result.WriteRune(' ')
				}
			}
		}

		result.WriteRune(r)
	}

	return result.String()
}

// processBrackets 处理各种类型的括号
func (tp *TextProcessor) processBrackets(text string) string {
	// 处理括号内容，为TTS添加适当的语调标记
	bracketPatterns := map[string][2]string{
		`（([^）]+）)`:    {"（", "）"},   // 中文括号
		`\(([^)]+)\)`:  {"(", ")"},   // 英文括号
		`【([^】]+】)`:    {"【", "】"},   // 中文方括号
		`\[([^\]]+)\]`: {"[", "]"},   // 英文方括号
		`《([^》]+》)`:    {"《", "》"},   // 中文书名号
		`"([^"]+")`:    {"\"", "\""}, // 中文双引号
		`'([^']+')`:    {"'", "'"},   // 中文单引号
		`"([^"]+)"`:    {"\"", "\""}, // 英文双引号
		`'([^']+)'`:    {"'", "'"},   // 英文单引号
	}

	for pattern := range bracketPatterns {
		regex := regexp.MustCompile(pattern)
		// 保持括号内容不变，只是确保括号周围有适当的停顿
		text = regex.ReplaceAllStringFunc(text, func(match string) string {
			return match // 保持原样，让TTS自然处理
		})
	}

	return text
}

// isChinese 判断是否为中文字符
func (tp *TextProcessor) isChinese(r rune) bool {
	return unicode.Is(unicode.Scripts["Han"], r)
}

// isEnglish 判断是否为英文字符
func (tp *TextProcessor) isEnglish(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

// IsValidTextForTTS 检查文本是否适合TTS处理
func (tp *TextProcessor) IsValidTextForTTS(text string) bool {
	text = strings.TrimSpace(text)

	// 空文本
	if text == "" {
		return false
	}

	// 检查是否为代码块
	if tp.isCodeBlock(text) {
		return false
	}

	// 检查是否为表格行
	if tp.isTableRow(text) || tp.isTableSeparator(text) {
		return false
	}

	// 检查是否为图片
	if tp.isImage(text) {
		return false
	}

	// 检查是否为纯URL或邮箱
	if tp.isPureURL(text) {
		return false
	}

	// 纯标记行（如 ###、**、-----）
	if tp.isPureMarkupLine(text) {
		return false
	}

	// 太短的文本（少于2个字符）
	if len([]rune(text)) < 2 {
		return false
	}

	// 检查是否包含有效内容（至少有一个字母、数字或中文字符）
	hasValidContent := false
	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || tp.isChinese(r) {
			hasValidContent = true
			break
		}
	}

	return hasValidContent
}

// isCodeBlock 检查是否为代码块
func (tp *TextProcessor) isCodeBlock(text string) bool {
	text = strings.TrimSpace(text)

	// 检查是否以代码块标记开始或结束
	if strings.HasPrefix(text, "```") || strings.HasSuffix(text, "```") {
		return true
	}

	if strings.HasPrefix(text, "~~~") || strings.HasSuffix(text, "~~~") {
		return true
	}

	// 检查是否为缩进代码（行首4个空格）
	if strings.HasPrefix(text, "    ") && len(strings.TrimLeft(text, " ")) > 0 {
		return true
	}

	// 检查常见的代码模式
	codePatterns := []string{
		`^func\s+\w+\s*\(`,         // Go函数定义
		`^package\s+\w+`,           // Go包声明
		`^import\s+`,               // 导入语句
		`^class\s+\w+`,             // 类定义
		`^def\s+\w+\s*\(`,          // Python函数定义
		`^if\s*\(.*\)\s*\{`,        // if语句
		`^for\s*\(.*\)\s*\{`,       // for循环 (C-style)
		`^for\s+\w+\s*:=.*\{`,      // Go for循环
		`^while\s*\(.*\)\s*\{`,     // while循环
		`^\s*\{`,                   // 单独的花括号
		`^\s*\}`,                   // 单独的花括号
		`^\s*return\s*;?\s*$`,      // return语句（修复：更严格的匹配）
		`^\s*return\s+[^a-zA-Z中文]`, // return带值
		`fmt\.Print`,               // 常见函数调用
		`console\.log`,             // JavaScript console
		`System\.out\.print`,       // Java输出
	}

	for _, pattern := range codePatterns {
		matched, _ := regexp.MatchString(pattern, text)
		if matched {
			return true
		}
	}

	return false
}

// isImage 检查是否为图片
func (tp *TextProcessor) isImage(text string) bool {
	text = strings.TrimSpace(text)

	// Markdown图片格式
	imageRegex := regexp.MustCompile(`^!\[([^\]]*)\]\([^)]+\)`)
	if imageRegex.MatchString(text) {
		return true
	}

	// HTML图片标签
	htmlImageRegex := regexp.MustCompile(`(?i)^<img[^>]*>`)
	if htmlImageRegex.MatchString(text) {
		return true
	}

	return false
}

// isPureURL 检查是否为纯URL或邮箱
func (tp *TextProcessor) isPureURL(text string) bool {
	text = strings.TrimSpace(text)

	// URL模式
	urlPatterns := []string{
		`^https?://[^\s]+$`,
		`^ftp://[^\s]+$`,
		`^www\.[^\s]+$`,
		`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`, // 邮箱
	}

	for _, pattern := range urlPatterns {
		if matched, _ := regexp.MatchString(pattern, text); matched {
			return true
		}
	}

	return false
}

// isPureMarkupLine 检查是否为纯标记行
func (tp *TextProcessor) isPureMarkupLine(text string) bool {
	text = strings.TrimSpace(text)

	// 检查各种标记格式
	markupPatterns := []string{
		`^#+\s*$`,                   // 纯井号
		`^\*+\s*$`,                  // 纯星号
		`^-+\s*$`,                   // 纯破折号
		`^=+\s*$`,                   // 纯等号
		`^_+\s*$`,                   // 纯下划线
		`^#+[^a-zA-Z\p{Han}]*$`,     // 井号加非字母内容
		`^\*{3,}[^a-zA-Z\p{Han}]*$`, // 三个或更多星号加非字母内容
		`^-{3,}[^a-zA-Z\p{Han}]*$`,  // 三个或更多破折号加非字母内容
		`^##.*$`,                    // 以 ## 开头的行（Markdown 标题）
		`^\*\*\(.*$`,                // 以 **( 开头的行（格式化说明）
		`^---.*$`,                   // 以 --- 开头的行（分割线）
		`^-----.*$`,                 // 以 ----- 开头的行（分割线）
		`^\|[-:|\\s]+\|$`,           // 表格分隔符行
		`^>\s*$`,                    // 空引用块
		`^[-*+]\s*$`,                // 空列表项
		`^\d+\.\s*$`,                // 空有序列表项
		`^[-*+]\s*\[[\sx]\]\s*$`,    // 空任务列表项
		`^\s*` + "`" + `{3}\s*$`,    // 代码块开始/结束标记
		`^\s*~{3}\s*$`,              // 代码块开始/结束标记（波浪号）
		`^<!--.*-->$`,               // HTML注释
		`^<[^>]+>\s*$`,              // 单独的HTML标签
	}

	for _, pattern := range markupPatterns {
		if matched, _ := regexp.MatchString(pattern, text); matched {
			return true
		}
	}

	return false
}

// SetOptions 设置处理器选项
func (tp *TextProcessor) SetOptions(preserveMarkdown, normalizeWhitespaceOpt, handleSpecialSymbols bool) {
	tp.preserveMarkdown = preserveMarkdown
	tp.normalizeWhitespace = normalizeWhitespaceOpt
	tp.handleSpecialSymbols = handleSpecialSymbols
}
