# TTS应用Makefile

# 项目信息
PROJECT_NAME := tts_app
VERSION ?= dev
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# 构建标志
LDFLAGS := -s -w -X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.gitCommit=$(GIT_COMMIT)

# 目录
DIST_DIR := dist
COVERAGE_DIR := coverage

# Go参数
GO := go
GOFMT := gofmt
GOFILES := $(shell find . -name "*.go" -type f)

# 默认目标
.DEFAULT_GOAL := help

# 帮助信息
.PHONY: help
help: ## 显示帮助信息
	@echo "TTS语音合成应用 - 构建工具"
	@echo ""
	@echo "可用命令:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# 开发相关
.PHONY: dev
dev: ## 开发模式构建（包含调试信息）
	@echo "构建开发版本..."
	$(GO) build -race -ldflags="$(LDFLAGS)" -o $(PROJECT_NAME) .

.PHONY: build
build: ## 构建当前平台的二进制文件
	@echo "构建发布版本..."
	CGO_ENABLED=0 $(GO) build -ldflags="$(LDFLAGS)" -o $(PROJECT_NAME) .

.PHONY: install
install: ## 安装到GOPATH/bin
	@echo "安装到 $(shell go env GOPATH)/bin/$(PROJECT_NAME)..."
	$(GO) install -ldflags="$(LDFLAGS)" .

# 测试相关
.PHONY: test
test: ## 运行测试
	@echo "运行测试..."
	$(GO) test -v ./...

.PHONY: test-coverage
test-coverage: ## 运行测试并生成覆盖率报告
	@echo "运行测试并生成覆盖率报告..."
	@mkdir -p $(COVERAGE_DIR)
	$(GO) test -v -coverprofile=$(COVERAGE_DIR)/coverage.out ./...
	$(GO) tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	@echo "覆盖率报告已生成: $(COVERAGE_DIR)/coverage.html"

.PHONY: test-race
test-race: ## 运行竞态检测测试
	@echo "运行竞态检测测试..."
	$(GO) test -race -v ./...

.PHONY: benchmark
benchmark: ## 运行基准测试
	@echo "运行基准测试..."
	$(GO) test -bench=. -benchmem ./...

# 代码质量
.PHONY: fmt
fmt: ## 格式化代码
	@echo "格式化代码..."
	$(GOFMT) -s -w $(GOFILES)

.PHONY: fmt-check
fmt-check: ## 检查代码格式
	@echo "检查代码格式..."
	@diff=$$($(GOFMT) -s -d $(GOFILES)); \
	if [ -n "$$diff" ]; then \
		echo "需要格式化的文件:"; \
		echo "$$diff"; \
		exit 1; \
	fi

.PHONY: vet
vet: ## 运行go vet
	@echo "运行 go vet..."
	$(GO) vet ./...

.PHONY: lint
lint: fmt-check vet ## 运行所有代码检查
	@echo "代码检查完成"

# 依赖管理
.PHONY: deps
deps: ## 下载依赖
	@echo "下载依赖..."
	$(GO) mod download

.PHONY: deps-update
deps-update: ## 更新依赖
	@echo "更新依赖..."
	$(GO) get -u ./...
	$(GO) mod tidy

.PHONY: deps-verify
deps-verify: ## 验证依赖
	@echo "验证依赖..."
	$(GO) mod verify

# 多平台构建
.PHONY: build-linux
build-linux: ## 构建Linux版本
	@echo "构建Linux版本..."
	@mkdir -p $(DIST_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/$(PROJECT_NAME)_linux_amd64 .
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GO) build -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/$(PROJECT_NAME)_linux_arm64 .

.PHONY: build-darwin
build-darwin: ## 构建macOS版本
	@echo "构建macOS版本..."
	@mkdir -p $(DIST_DIR)
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GO) build -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/$(PROJECT_NAME)_darwin_amd64 .
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(GO) build -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/$(PROJECT_NAME)_darwin_arm64 .

.PHONY: build-windows
build-windows: ## 构建Windows版本
	@echo "构建Windows版本..."
	@mkdir -p $(DIST_DIR)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GO) build -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/$(PROJECT_NAME)_windows_amd64.exe .

.PHONY: build-all
build-all: build-linux build-darwin build-windows ## 构建所有平台
	@echo "所有平台构建完成"
	@ls -la $(DIST_DIR)/

# 打包
.PHONY: package
package: build-all ## 打包发布文件
	@echo "创建发布包..."
	@cd $(DIST_DIR) && \
	for file in $(PROJECT_NAME)_linux_*; do \
		tar -czf "$$file.tar.gz" "$$file" ../config.yaml.example ../README.md ../LICENSE; \
	done && \
	for file in $(PROJECT_NAME)_darwin_*; do \
		tar -czf "$$file.tar.gz" "$$file" ../config.yaml.example ../README.md ../LICENSE; \
	done && \
	for file in $(PROJECT_NAME)_windows_*.exe; do \
		zip "$${file%.exe}.zip" "$$file" ../config.yaml.example ../README.md ../LICENSE; \
	done
	@echo "发布包创建完成:"
	@find $(DIST_DIR) -name "*.tar.gz" -o -name "*.zip" | sort

# Docker
.PHONY: docker-build
docker-build: ## 构建Docker镜像
	@echo "构建Docker镜像..."
	docker build -t $(PROJECT_NAME):$(VERSION) .
	docker build -t $(PROJECT_NAME):latest .

.PHONY: docker-run
docker-run: ## 运行Docker容器
	@echo "运行Docker容器..."
	docker run --rm -it \
		-v $(PWD)/test.txt:/app/input.txt:ro \
		-v $(PWD)/output:/app/output \
		$(PROJECT_NAME):latest edge -i input.txt

.PHONY: docker-push
docker-push: ## 推送Docker镜像
	@echo "推送Docker镜像..."
	docker push $(PROJECT_NAME):$(VERSION)
	docker push $(PROJECT_NAME):latest

# 发布
.PHONY: release
release: clean lint test package ## 完整发布流程
	@echo "发布完成 (版本: $(VERSION))"

# 清理
.PHONY: clean
clean: ## 清理构建产物
	@echo "清理构建产物..."
	@rm -rf $(DIST_DIR)/
	@rm -rf $(COVERAGE_DIR)/
	@rm -f $(PROJECT_NAME)
	@rm -f coverage.out
	@rm -f *.exe

.PHONY: clean-deps
clean-deps: ## 清理依赖缓存
	@echo "清理依赖缓存..."
	$(GO) clean -modcache

# 开发工具
.PHONY: dev-tools
dev-tools: ## 安装开发工具
	@echo "安装开发工具..."
	$(GO) install golang.org/x/tools/cmd/goimports@latest
	$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GO) install github.com/securecodewarrior/sast-scan@latest

# 调试和运行
.PHONY: run
run: build ## 构建并运行（显示帮助）
	@echo "运行应用..."
	./$(PROJECT_NAME) --help

.PHONY: run-edge
run-edge: build ## 构建并运行Edge TTS示例
	@echo "运行Edge TTS示例..."
	@echo "欢迎使用TTS应用" > test_input.txt
	./$(PROJECT_NAME) edge -i test_input.txt
	@echo "输出文件: output/merged_audio.mp3"

# 信息查看
.PHONY: info
info: ## 显示构建信息
	@echo "项目信息:"
	@echo "  名称: $(PROJECT_NAME)"
	@echo "  版本: $(VERSION)"
	@echo "  构建时间: $(BUILD_TIME)"
	@echo "  Git提交: $(GIT_COMMIT)"
	@echo "  Go版本: $(shell $(GO) version)"
	@echo "  LDFLAGS: $(LDFLAGS)"

.PHONY: deps-graph
deps-graph: ## 显示依赖图
	@echo "依赖图:"
	$(GO) mod graph

# 检查更新
.PHONY: outdated
outdated: ## 检查过时的依赖
	@echo "检查过时的依赖..."
	$(GO) list -u -m all

# 安全扫描
.PHONY: security
security: ## 运行安全扫描
	@echo "运行安全扫描..."
	$(GO) list -json -m all | nancy sleuth

# 性能分析
.PHONY: profile-cpu
profile-cpu: ## CPU性能分析
	@echo "运行CPU性能分析..."
	$(GO) test -cpuprofile=cpu.prof -bench=. ./...

.PHONY: profile-mem
profile-mem: ## 内存性能分析
	@echo "运行内存性能分析..."
	$(GO) test -memprofile=mem.prof -bench=. ./...

# 文档生成
.PHONY: docs
docs: ## 生成文档
	@echo "生成文档..."
	$(GO) doc -all . > docs/api.md

# 验证
.PHONY: verify
verify: deps-verify lint test ## 完整验证流程
	@echo "验证完成"
