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
}

// NewTextProcessor 创建新的文本处理器
func NewTextProcessor() *TextProcessor {
	return &TextProcessor{
		preserveMarkdown:     true,
		normalizeWhitespace:  true,
		handleSpecialSymbols: true,
	}
}

// ProcessText 处理文本，优化TTS语音合成效果
func (tp *TextProcessor) ProcessText(text string) string {
	if text == "" {
		return text
	}

	// 1. 处理转义字符（需要在Markdown处理之前）
	text = tp.processEscapeCharacters(text)

	// 2. 处理Markdown格式字符
	if tp.preserveMarkdown {
		text = tp.processMarkdownFormatting(text)
	}

	// 3. 处理特殊符号
	if tp.handleSpecialSymbols {
		text = tp.processSpecialSymbols(text)
	}

	// 4. 规范化空白字符
	if tp.normalizeWhitespace {
		text = tp.normalizeWhitespaceText(text)
	}

	// 5. 处理中英文混合文本
	text = tp.processMixedLanguageText(text)

	// 6. 处理各种类型的括号
	text = tp.processBrackets(text)

	return text
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
		`\*`:  "*",   // 转义的星号
		`\\`:  "\\",  // 转义的反斜杠
		`\n`:  " ",   // 换行符转为空格
		`\t`:  " ",   // 制表符转为空格
		`\r`:  "",    // 回车符删除
		`\"`:  "\"",  // 转义的双引号
		`\'`:  "'",   // 转义的单引号
		`\&`:  "&",   // 转义的&符号
		`\#`:  "#",   // 转义的#符号
		`\%`:  "%",   // 转义的%符号
		`\$`:  "$",   // 转义的$符号
		`\@`:  "@",   // 转义的@符号
		`\!`:  "!",   // 转义的!符号
		`\?`:  "?",   // 转义的?符号
		`\+`:  "+",   // 转义的+符号
		`\=`:  "=",   // 转义的=符号
		`\-`:  "-",   // 转义的-符号
		`\_`:  "_",   // 转义的下划线
		`\^`:  "^",   // 转义的^符号
		`\~`:  "~",   // 转义的~符号
		`\|`:  "|",   // 转义的|符号
		`\>`:  ">",   // 转义的>符号
		`\<`:  "<",   // 转义的<符号
		`\{`:  "{",   // 转义的{符号
		`\}`:  "}",   // 转义的}符号
		`\[`:  "[",   // 转义的[符号
		`\]`:  "]",   // 转义的]符号
		`\(`:  "(",   // 转义的(符号
		`\)`:  ")",   // 转义的)符号
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
		"@":  "at",
		"#":  "井号", 
		"$":  "美元",
		"%":  "百分号",
		"^":  "上标",
		"&":  "和",
		"*":  "星号",
		"+":  "加号",
		"=":  "等号",
		"|":  "竖线",
		"~":  "波浪号",
		"`":  "反引号",
		"<":  "小于号",
		">":  "大于号",
		"[":  "左方括号",
		"]":  "右方括号",
		"{":  "左大括号",
		"}":  "右大括号",
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
		`\w+@\w+\.\w+`,           // 邮箱地址
		`https?://[^\s]+`,        // 网址
		`\$\d+`,                  // 价格（美元）
		`\d+%`,                   // 百分比
		`\d+\.\d+`,               // 小数
		`#[a-zA-Z_]\w*`,          // 编程中的标识符
		`\*+[^*]*\*+`,            // 被星号包围的文本
		`\+\d+(-\d+)*`,           // 电话号码
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
		`（([^）]+）)`: {"（", "）"}, // 中文括号
		`\(([^)]+)\)`: {"(", ")"},   // 英文括号
		`【([^】]+】)`: {"【", "】"}, // 中文方括号
		`\[([^\]]+)\]`: {"[", "]"}, // 英文方括号
		`《([^》]+》)`: {"《", "》"}, // 中文书名号
		`"([^"]+")`:   {"\"", "\""}, // 中文双引号
		`'([^']+')`:   {"'", "'"}, // 中文单引号
		`"([^"]+)"`:   {"\"", "\""}, // 英文双引号
		`'([^']+)'`:   {"'", "'"}, // 英文单引号
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

// isPureMarkupLine 检查是否为纯标记行
func (tp *TextProcessor) isPureMarkupLine(text string) bool {
	text = strings.TrimSpace(text)
	
	// 检查各种标记格式
	markupPatterns := []string{
		`^#+\s*$`,           // 纯井号
		`^\*+\s*$`,          // 纯星号
		`^-+\s*$`,           // 纯破折号
		`^=+\s*$`,           // 纯等号
		`^_+\s*$`,           // 纯下划线
		`^#+[^a-zA-Z\p{Han}]*$`, // 井号加非字母内容
		`^\*{3,}[^a-zA-Z\p{Han}]*$`, // 三个或更多星号加非字母内容
		`^-{3,}[^a-zA-Z\p{Han}]*$`,  // 三个或更多破折号加非字母内容
		`^##.*$`,            // 以 ## 开头的行（Markdown 标题）
		`^\*\*\(.*$`,        // 以 **( 开头的行（格式化说明）
		`^---.*$`,           // 以 --- 开头的行（分割线）
		`^-----.*$`,         // 以 ----- 开头的行（分割线）
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
