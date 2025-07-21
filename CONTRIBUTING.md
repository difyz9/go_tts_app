# 贡献指南

感谢您对 TTS语音合成应用 项目的关注和贡献！

## 🤝 如何贡献

### 贡献类型

我们欢迎以下类型的贡献：

- 🐛 **Bug 报告和修复**
- ✨ **新功能开发**
- 📚 **文档改进**
- 🎨 **UI/UX 改进**
- ⚡ **性能优化**
- 🧪 **测试覆盖**
- 🌍 **国际化支持**

### 开发流程

1. **Fork 项目**
   ```bash
   # 在GitHub上点击Fork按钮
   # 然后克隆你的fork
   git clone https://github.com/difyz9/markdown2tts.git
   cd go-tts-app
   ```

2. **设置开发环境**
   ```bash
   # 安装Go 1.23+
   # 安装依赖
   go mod download
   
   # 运行测试确保环境正常
   go test ./...
   ```

3. **创建功能分支**
   ```bash
   git checkout -b feature/your-feature-name
   # 或者
   git checkout -b fix/your-bug-fix
   ```

4. **进行开发**
   - 遵循Go编码规范
   - 添加必要的测试
   - 更新相关文档

5. **提交代码**
   ```bash
   git add .
   git commit -m "feat: add awesome feature"
   # 或者
   git commit -m "fix: resolve memory leak issue"
   ```

6. **推送并创建PR**
   ```bash
   git push origin feature/your-feature-name
   # 在GitHub上创建Pull Request
   ```

## 📝 提交规范

我们使用 [Conventional Commits](https://www.conventionalcommits.org/) 规范：

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

### 提交类型

- `feat`: 新功能
- `fix`: Bug修复
- `docs`: 文档更新
- `style`: 代码格式（不影响功能）
- `refactor`: 重构
- `perf`: 性能优化
- `test`: 测试相关
- `chore`: 构建/工具相关

### 示例

```bash
feat(edge-tts): add support for new voice models
fix(concurrent): resolve race condition in worker pool
docs(readme): update installation instructions
perf(audio): optimize memory usage in audio processing
```

## 🧪 测试

### 运行测试

```bash
# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./service

# 运行测试并查看覆盖率
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 编写测试

- 为新功能添加单元测试
- 确保测试覆盖率不降低
- 使用表驱动测试模式
- 模拟外部依赖

示例测试：

```go
func TestEdgeTTSService_ProcessText(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {
            name:    "valid text",
            input:   "Hello world",
            want:    "audio_file.mp3",
            wantErr: false,
        },
        // 更多测试用例...
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // 测试逻辑
        })
    }
}
```

## 📚 代码规范

### Go代码规范

1. **使用gofmt格式化代码**
   ```bash
   gofmt -s -w .
   ```

2. **使用golint检查代码**
   ```bash
   golint ./...
   ```

3. **使用go vet检查**
   ```bash
   go vet ./...
   ```

4. **遵循Go最佳实践**
   - 使用有意义的变量名
   - 添加适当的注释
   - 处理所有错误
   - 使用context进行超时控制

### 代码结构

```
项目结构遵循标准Go项目布局：

cmd/        # 命令行接口
model/      # 数据模型
service/    # 业务逻辑
internal/   # 内部包（如果需要）
pkg/        # 可重用的包（如果需要）
test/       # 测试数据和工具
docs/       # 文档
```

## 🐛 报告Bug

### Bug报告模板

请使用以下模板报告Bug：

```markdown
## Bug描述
简要描述遇到的问题

## 复现步骤
1. 执行命令 `...`
2. 输入参数 `...`
3. 看到错误 `...`

## 预期行为
描述您期望的正确行为

## 实际行为
描述实际发生的情况

## 环境信息
- OS: [e.g., Ubuntu 20.04]
- Go版本: [e.g., 1.23.0]
- 应用版本: [e.g., v1.0.0]

## 附加信息
- 错误日志
- 配置文件内容
- 屏幕截图（如适用）
```

## ✨ 功能请求

### 功能请求模板

```markdown
## 功能描述
简要描述您希望添加的功能

## 使用场景
描述什么情况下需要这个功能

## 期望的行为
详细描述功能应该如何工作

## 可能的实现方案
如果有想法，可以描述可能的实现方式

## 相关资源
提供相关的链接、文档或参考资料
```

## 📖 文档贡献

### 文档类型

- **README.md**: 项目介绍和快速开始
- **API文档**: 代码中的注释文档
- **使用教程**: 详细的使用指南
- **开发文档**: 开发相关的说明

### 文档规范

1. 使用Markdown格式
2. 添加适当的代码示例
3. 保持简洁明了
4. 及时更新过时内容

## 🚀 发布流程

### 版本管理

我们使用语义化版本控制（SemVer）：

- `MAJOR.MINOR.PATCH`
- 主版本号：不兼容的API修改
- 次版本号：向下兼容的功能性新增
- 修订号：向下兼容的问题修正

### 发布步骤

1. 更新版本号
2. 更新CHANGELOG.md
3. 创建Git标签
4. GitHub Actions自动构建和发布

## 🤗 社区

### 交流渠道

- 🐛 [GitHub Issues](https://github.com/difyz9/markdown2tts/issues) - Bug报告和功能请求
- 💬 [GitHub Discussions](https://github.com/difyz9/markdown2tts/discussions) - 一般讨论
- 📧 Email: your-email@example.com

### 行为准则

我们致力于为每个人提供友好、安全和欢迎的环境：

- 使用友好和包容的语言
- 尊重不同的观点和经验
- 优雅地接受建设性批评
- 专注于对社区最有利的事情
- 对其他社区成员表示同理心

## 🙏 致谢

感谢所有为这个项目做出贡献的开发者！

### 贡献者列表

- [@contributor1](https://github.com/contributor1) - 功能开发
- [@contributor2](https://github.com/contributor2) - Bug修复
- [@contributor3](https://github.com/contributor3) - 文档改进

---

再次感谢您的贡献！如果您有任何问题，请随时通过Issue或Discussion联系我们。
