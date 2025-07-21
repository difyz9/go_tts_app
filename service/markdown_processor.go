package service

import (
	"bytes"
	"regexp"
	"strings"

	"github.com/russross/blackfriday/v2"
)

// MarkdownProcessor 专门处理Markdown文档的处理器
type MarkdownProcessor struct {
	preserveLinks bool
	removeImages  bool
}

// NewMarkdownProcessor 创建新的Markdown处理器
func NewMarkdownProcessor() *MarkdownProcessor {
	return &MarkdownProcessor{
		preserveLinks: true, // 保留链接文本
		removeImages:  true, // 移除图片
	}
}

// ExtractTextForTTS 从Markdown文档中提取适合TTS的纯文本
func (mp *MarkdownProcessor) ExtractTextForTTS(markdown string) string {
	// 使用 blackfriday 解析 Markdown
	doc := blackfriday.New(blackfriday.WithExtensions(
		blackfriday.CommonExtensions |
			blackfriday.AutoHeadingIDs |
			blackfriday.Footnotes,
	)).Parse([]byte(markdown))

	// 创建自定义渲染器来提取纯文本
	renderer := &TTSRenderer{
		preserveLinks: mp.preserveLinks,
		removeImages:  mp.removeImages,
		buffer:        &bytes.Buffer{},
	}

	// 遍历AST并提取文本
	doc.Walk(func(node *blackfriday.Node, entering bool) blackfriday.WalkStatus {
		return renderer.RenderNode(node, entering)
	})

	result := renderer.buffer.String()

	// 后处理：清理多余的空白字符
	result = mp.cleanupText(result)

	return result
}

// TTSRenderer 自定义渲染器，专门用于提取适合TTS的文本
type TTSRenderer struct {
	preserveLinks bool
	removeImages  bool
	buffer        *bytes.Buffer
	inImage       bool
	linkText      string
}

// RenderNode 处理AST节点
func (r *TTSRenderer) RenderNode(node *blackfriday.Node, entering bool) blackfriday.WalkStatus {
	switch node.Type {
	case blackfriday.CodeBlock:
		// 完全跳过代码块，但不影响后续节点的处理
		return blackfriday.SkipChildren

	case blackfriday.Code:
		// 保留内联代码内容（但移除反引号标记）
		// 内联代码通常是技术术语，对TTS有价值
		if entering && node.Literal != nil {
			text := string(node.Literal)
			r.buffer.WriteString(text)
			r.buffer.WriteString(" ")
		}
		return blackfriday.SkipChildren

	case blackfriday.HTMLBlock, blackfriday.HTMLSpan:
		// 跳过HTML块，但可能需要提取内容
		if entering && r.shouldExtractHTMLContent(node) {
			content := r.extractHTMLContent(string(node.Literal))
			if content != "" {
				r.buffer.WriteString(content)
				r.buffer.WriteString(" ")
			}
		}
		return blackfriday.SkipChildren

	case blackfriday.Image:
		// 处理图片
		if r.removeImages {
			return blackfriday.SkipChildren
		}
		if entering {
			r.inImage = true
		} else {
			r.inImage = false
		}
		return blackfriday.SkipChildren

	case blackfriday.Link:
		// 处理链接
		if entering {
			r.linkText = ""
		} else {
			if r.preserveLinks && r.linkText != "" {
				r.buffer.WriteString(r.linkText)
				r.buffer.WriteString(" ")
			}
		}
		return blackfriday.GoToNext

	case blackfriday.Text:
		// 处理文本节点
		if !r.inImage {
			text := string(node.Literal)

			// 如果在链接中，收集链接文本
			if node.Parent != nil && node.Parent.Type == blackfriday.Link {
				r.linkText += text
			} else {
				// 普通文本，直接添加
				r.buffer.WriteString(text)
				r.buffer.WriteString(" ")
			}
		}

	case blackfriday.Heading:
		// 跳过所有级别的标题（H1-H6）
		return blackfriday.SkipChildren

	case blackfriday.Paragraph:
		// 段落处理
		if !entering {
			r.buffer.WriteString("\n")
		}

	case blackfriday.List, blackfriday.Item:
		// 列表处理
		if !entering {
			r.buffer.WriteString("\n")
		}

	case blackfriday.BlockQuote:
		// 引用块处理
		if !entering {
			r.buffer.WriteString("\n")
		}

	case blackfriday.Table, blackfriday.TableHead, blackfriday.TableBody, blackfriday.TableRow, blackfriday.TableCell:
		// 跳过表格
		return blackfriday.SkipChildren
	}

	return blackfriday.GoToNext
}

// shouldExtractHTMLContent 判断是否应该提取HTML内容
func (r *TTSRenderer) shouldExtractHTMLContent(node *blackfriday.Node) bool {
	content := string(node.Literal)

	// 跳过脚本和样式标签
	if strings.Contains(content, "<script") ||
		strings.Contains(content, "<style") ||
		strings.Contains(content, "<img") {
		return false
	}

	return true
}

// extractHTMLContent 从HTML中提取文本内容
func (r *TTSRenderer) extractHTMLContent(html string) string {
	// 简单的HTML标签移除
	tagRegex := regexp.MustCompile(`<[^>]*>`)
	content := tagRegex.ReplaceAllString(html, " ")

	// 处理HTML实体
	content = strings.ReplaceAll(content, "&nbsp;", " ")
	content = strings.ReplaceAll(content, "&amp;", "&")
	content = strings.ReplaceAll(content, "&lt;", "<")
	content = strings.ReplaceAll(content, "&gt;", ">")
	content = strings.ReplaceAll(content, "&quot;", "\"")
	content = strings.ReplaceAll(content, "&#39;", "'")

	return strings.TrimSpace(content)
}

// cleanupText 清理文本中的多余空白字符
func (mp *MarkdownProcessor) cleanupText(text string) string {
	// 移除多余的空白字符
	spaceRegex := regexp.MustCompile(`\s+`)
	text = spaceRegex.ReplaceAllString(text, " ")

	// 移除多余的换行符
	newlineRegex := regexp.MustCompile(`\n\s*\n`)
	text = newlineRegex.ReplaceAllString(text, "\n")

	// 移除开头和结尾的空白
	text = strings.TrimSpace(text)

	return text
}

// SplitIntoSentences 将文本分割成适合TTS的句子
func (mp *MarkdownProcessor) SplitIntoSentences(text string) []string {
	if text == "" {
		return []string{}
	}

	// 按换行符分割段落
	paragraphs := strings.Split(text, "\n")
	var sentences []string

	for _, paragraph := range paragraphs {
		paragraph = strings.TrimSpace(paragraph)
		if paragraph == "" {
			continue
		}

		// 保护常见的技术术语，避免在其中分割
		protected := paragraph

		// 暂时替换常见的技术模式，避免在这些地方分割
		protectedPatterns := map[string]string{
			".New()":  "NEWMETHOD",
			".Load()": "LOADMETHOD",
			".Call()": "CALLMETHOD",
			".com/":   "DOTCOM",
			".org/":   "DOTORG",
			".net/":   "DOTNET",
			".go":     "DOTGO",
		}

		for pattern, replacement := range protectedPatterns {
			protected = strings.ReplaceAll(protected, pattern, replacement)
		}

		// 现在可以安全地按句号分割（只对中文句号和英文句号结尾）
		sentenceRegex := regexp.MustCompile(`[。！？]|[.!?](?:\s|$)`)
		if sentenceRegex.MatchString(protected) {
			parts := sentenceRegex.Split(protected, -1)
			matches := sentenceRegex.FindAllString(protected, -1)

			for i, part := range parts {
				part = strings.TrimSpace(part)
				if part == "" {
					continue
				}

				// 恢复保护的模式
				for pattern, replacement := range protectedPatterns {
					part = strings.ReplaceAll(part, replacement, pattern)
				}

				// 加回标点符号（除了最后一部分）
				if i < len(matches) {
					part += matches[i]
				}

				sentences = append(sentences, part)
			}
		} else {
			// 恢复保护的模式
			for pattern, replacement := range protectedPatterns {
				paragraph = strings.ReplaceAll(paragraph, replacement, pattern)
			}
			sentences = append(sentences, paragraph)
		}
	}

	return sentences
}
