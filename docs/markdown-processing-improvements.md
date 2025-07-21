# Markdown文本处理改进总结

## 更新时间
2025年7月21日

## 改进目标
增强TTS应用的文本处理能力，使其能够智能过滤Markdown文档中不适合语音合成的内容，包括：
- 代码块
- 表格
- 图片
- 链接（保留文本部分）
- HTML标签

## 主要改进

### 1. 新增removeNonSpeechElements方法
在文本处理的第一步就移除不需要语音合成的元素：

```go
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
```

### 2. 代码块处理
- **多行代码块**：移除```或~~~包围的代码块
- **缩进代码块**：移除行首4个空格的代码行
- **智能识别**：在IsValidTextForTTS中识别代码块标记行

#### 处理效果：
```markdown
原文：
```go
func main() {
    fmt.Println("Hello")
}
```

处理结果：完全移除，不进行语音合成
```

### 3. 表格处理
- **表格行识别**：检测包含管道符|的表格行
- **表格分隔符**：识别|---|---|格式的分隔符
- **完整表格移除**：从表格开始到结束完整移除

#### 处理效果：
```markdown
原文：
| 列1 | 列2 | 列3 |
|-----|-----|-----|
| 数据1 | 数据2 | 数据3 |

处理结果：完全移除表格内容
```

### 4. 图片处理
- **Markdown图片**：移除![alt](url)格式
- **HTML图片**：移除<img>标签
- **智能识别**：在文本验证阶段提前识别

#### 处理效果：
```markdown
原文：![示例图片](https://example.com/image.jpg)
处理结果：完全移除
```

### 5. 链接处理
- **保留文本**：从[text](url)中提取text部分
- **移除URL**：删除纯http://、https://、www.等链接
- **移除邮箱**：删除邮箱地址

#### 处理效果：
```markdown
原文：这是[百度](https://www.baidu.com)链接
处理结果：这是百度链接

原文：访问 https://www.example.com
处理结果：访问

原文：联系 test@example.com
处理结果：联系
```

### 6. HTML标签处理
- **移除标签**：删除所有HTML标签但保留内容
- **HTML实体**：转换常见HTML实体为对应字符

#### 处理效果：
```markdown
原文：<div>这是HTML标签中的内容</div>
处理结果：这是HTML标签中的内容

原文：&copy; 版权所有
处理结果：版权 版权所有
```

### 7. 增强的文本验证
更新IsValidTextForTTS方法，新增以下检查：
- 代码块检查
- 表格行检查
- 图片检查
- 纯URL检查
- 扩展的标记行检查

### 8. 测试验证结果

根据测试输出，系统能够正确：
- ✅ 跳过空行和标记行
- ✅ 完全移除代码块内容
- ✅ 移除表格和表格分隔符
- ✅ 移除图片引用
- ✅ 保留链接文本，移除URL
- ✅ 移除纯URL和邮箱
- ✅ 移除HTML标签但保留内容
- ✅ 保留有意义的文本内容

## 使用场景对比

### 改进前
```
输入：包含大量代码、表格、图片的Markdown文档
输出：所有内容都会尝试进行语音合成，包括：
- 代码变量名和语法
- 表格分隔符和数据
- 图片URL
- 链接URL
```

### 改进后
```
输入：包含大量代码、表格、图片的Markdown文档
输出：只对有意义的文本内容进行语音合成：
- 标题和正文
- 列表内容（移除标记符号）
- 链接的描述文本
- 引用块内容
```

## 配置说明

文本处理器支持以下配置选项：
- `preserveMarkdown`: 是否保留Markdown格式处理
- `normalizeWhitespace`: 是否规范化空白字符
- `handleSpecialSymbols`: 是否处理特殊符号

## 注意事项

1. **代码块识别**：基于```、~~~和缩进规则
2. **表格识别**：依赖管道符|的数量和模式
3. **链接处理**：会保留链接文本但移除URL
4. **HTML处理**：移除标签但保留内容
5. **性能影响**：增加了正则表达式处理，但对整体性能影响较小

## 兼容性

- 兼容CommonMark标准
- 支持GitHub Flavored Markdown
- 支持基本HTML标签
- 向后兼容现有配置

这些改进让TTS应用能够更智能地处理技术文档、博客文章和其他包含复杂Markdown格式的内容，确保语音合成的内容更加有意义和易于理解。
