#!/bin/bash

# TTS应用构建脚本
# 支持跨平台编译和打包

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 项目信息
PROJECT_NAME="github.com/difyz9/markdown2tts"
VERSION=${VERSION:-"dev"}
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# 构建标志
LDFLAGS="-s -w -X main.version=${VERSION} -X main.buildTime=${BUILD_TIME} -X main.gitCommit=${GIT_COMMIT}"

# 支持的平台
PLATFORMS=(
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
    "windows/amd64"
)

# 函数定义
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 显示帮助信息
show_help() {
    cat << EOF
TTS应用构建脚本

用法:
    $0 [选项] [命令]

命令:
    build       构建当前平台的二进制文件
    build-all   构建所有平台的二进制文件
    clean       清理构建产物
    test        运行测试
    lint        代码检查
    dev         开发模式构建（包含调试信息）
    release     发布模式构建并打包
    docker      构建Docker镜像
    help        显示帮助信息

选项:
    -v, --version VERSION   指定版本号 (默认: dev)
    -o, --output DIR        指定输出目录 (默认: dist)
    --race                  启用竞态检测
    --no-cgo                禁用CGO
    -h, --help              显示帮助信息

示例:
    $0 build                      # 构建当前平台
    $0 build-all                  # 构建所有平台
    $0 -v v1.0.0 release          # 发布v1.0.0版本
    $0 test                       # 运行测试
    $0 docker                     # 构建Docker镜像

EOF
}

# 检查依赖
check_dependencies() {
    log_info "检查构建依赖..."
    
    if ! command -v go &> /dev/null; then
        log_error "Go 未安装或不在PATH中"
        exit 1
    fi
    
    GO_VERSION=$(go version | grep -o 'go[0-9]\+\.[0-9]\+' | sed 's/go//')
    REQUIRED_VERSION="1.23"
    
    if ! printf '%s\n%s\n' "$REQUIRED_VERSION" "$GO_VERSION" | sort -V -C; then
        log_error "需要Go $REQUIRED_VERSION或更高版本，当前版本: $GO_VERSION"
        exit 1
    fi
    
    log_success "Go版本检查通过: $GO_VERSION"
}

# 清理构建产物
clean() {
    log_info "清理构建产物..."
    rm -rf dist/
    rm -rf coverage.out
    rm -rf *.exe
    rm -rf ${PROJECT_NAME}_*
    log_success "清理完成"
}

# 运行测试
run_tests() {
    log_info "运行测试..."
    
    # 运行单元测试
    go test -v ./... -coverprofile=coverage.out
    
    # 显示覆盖率
    if [ -f coverage.out ]; then
        COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
        log_success "测试完成，代码覆盖率: $COVERAGE"
        
        # 生成HTML覆盖率报告
        go tool cover -html=coverage.out -o coverage.html
        log_info "覆盖率报告已生成: coverage.html"
    fi
}

# 代码检查
run_lint() {
    log_info "运行代码检查..."
    
    # go vet
    log_info "运行 go vet..."
    go vet ./...
    
    # go fmt检查
    log_info "检查代码格式..."
    UNFORMATTED=$(gofmt -s -l .)
    if [ -n "$UNFORMATTED" ]; then
        log_error "以下文件需要格式化:"
        echo "$UNFORMATTED"
        exit 1
    fi
    
    # 如果安装了golangci-lint，运行它
    if command -v golangci-lint &> /dev/null; then
        log_info "运行 golangci-lint..."
        golangci-lint run
    else
        log_warning "golangci-lint 未安装，跳过高级检查"
    fi
    
    log_success "代码检查通过"
}

# 构建单个平台
build_platform() {
    local platform=$1
    local output_dir=$2
    local dev_mode=${3:-false}
    
    IFS='/' read -r GOOS GOARCH <<< "$platform"
    
    local binary_name="${PROJECT_NAME}_${GOOS}_${GOARCH}"
    if [ "$GOOS" = "windows" ]; then
        binary_name="${binary_name}.exe"
    fi
    
    local binary_path="${output_dir}/${binary_name}"
    
    log_info "构建 $platform -> $binary_path"
    
    # 设置构建标志
    local build_flags=""
    if [ "$dev_mode" = "true" ]; then
        build_flags="-race"
        LDFLAGS="-X main.version=${VERSION}-dev -X main.buildTime=${BUILD_TIME} -X main.gitCommit=${GIT_COMMIT}"
    fi
    
    # 构建
    env GOOS=$GOOS GOARCH=$GOARCH CGO_ENABLED=0 \
        go build $build_flags -ldflags="$LDFLAGS" -o "$binary_path" .
    
    if [ $? -eq 0 ]; then
        log_success "构建成功: $binary_path"
        
        # 显示文件信息
        if command -v ls &> /dev/null; then
            ls -lh "$binary_path"
        fi
    else
        log_error "构建失败: $platform"
        exit 1
    fi
}

# 构建当前平台
build_current() {
    local output_dir=${OUTPUT_DIR:-"dist"}
    local dev_mode=${1:-false}
    
    mkdir -p "$output_dir"
    
    check_dependencies
    
    CURRENT_PLATFORM="$(go env GOOS)/$(go env GOARCH)"
    build_platform "$CURRENT_PLATFORM" "$output_dir" "$dev_mode"
}

# 构建所有平台
build_all() {
    local output_dir=${OUTPUT_DIR:-"dist"}
    
    mkdir -p "$output_dir"
    
    check_dependencies
    
    log_info "构建所有平台..."
    
    for platform in "${PLATFORMS[@]}"; do
        build_platform "$platform" "$output_dir"
    done
    
    log_success "所有平台构建完成"
    log_info "构建产物位于: $output_dir/"
    ls -la "$output_dir/"
}

# 发布模式构建
build_release() {
    local output_dir=${OUTPUT_DIR:-"dist"}
    
    log_info "发布模式构建 (版本: $VERSION)..."
    
    # 清理
    clean
    
    # 运行测试
    run_tests
    
    # 代码检查
    run_lint
    
    # 构建所有平台
    build_all
    
    # 创建发布包
    log_info "创建发布包..."
    
    cd "$output_dir"
    
    for platform in "${PLATFORMS[@]}"; do
        IFS='/' read -r GOOS GOARCH <<< "$platform"
        
        local binary_name="${PROJECT_NAME}_${GOOS}_${GOARCH}"
        if [ "$GOOS" = "windows" ]; then
            binary_name="${binary_name}.exe"
            # Windows ZIP包
            zip "${binary_name%.exe}.zip" "$binary_name" ../config.yaml.example ../README.md ../LICENSE
        else
            # Linux/macOS tar.gz包
            tar -czf "${binary_name}.tar.gz" "$binary_name" ../config.yaml.example ../README.md ../LICENSE
        fi
    done
    
    cd ..
    
    log_success "发布包创建完成"
    log_info "发布文件:"
    find "$output_dir" -name "*.tar.gz" -o -name "*.zip" | sort
}

# 构建Docker镜像
build_docker() {
    log_info "构建Docker镜像..."
    
    local image_name="go-tts-app"
    local tag=${VERSION:-"latest"}
    
    docker build -t "${image_name}:${tag}" .
    
    if [ $? -eq 0 ]; then
        log_success "Docker镜像构建成功: ${image_name}:${tag}"
        
        # 显示镜像信息
        docker images | grep "$image_name"
        
        log_info "运行Docker容器:"
        echo "docker run --rm -v \$(pwd)/input.txt:/app/input.txt -v \$(pwd)/output:/app/output ${image_name}:${tag} edge -i input.txt"
    else
        log_error "Docker镜像构建失败"
        exit 1
    fi
}

# 解析命令行参数
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -v|--version)
                VERSION="$2"
                shift 2
                ;;
            -o|--output)
                OUTPUT_DIR="$2"
                shift 2
                ;;
            --race)
                RACE_ENABLED=true
                shift
                ;;
            --no-cgo)
                export CGO_ENABLED=0
                shift
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            build)
                COMMAND="build"
                shift
                ;;
            build-all)
                COMMAND="build-all"
                shift
                ;;
            clean)
                COMMAND="clean"
                shift
                ;;
            test)
                COMMAND="test"
                shift
                ;;
            lint)
                COMMAND="lint"
                shift
                ;;
            dev)
                COMMAND="dev"
                shift
                ;;
            release)
                COMMAND="release"
                shift
                ;;
            docker)
                COMMAND="docker"
                shift
                ;;
            help)
                show_help
                exit 0
                ;;
            *)
                log_error "未知参数: $1"
                show_help
                exit 1
                ;;
        esac
    done
}

# 主函数
main() {
    # 显示构建信息
    log_info "TTS应用构建脚本"
    log_info "版本: $VERSION"
    log_info "构建时间: $BUILD_TIME"
    log_info "Git提交: $GIT_COMMIT"
    echo
    
    # 执行命令
    case ${COMMAND:-"help"} in
        build)
            build_current
            ;;
        build-all)
            build_all
            ;;
        clean)
            clean
            ;;
        test)
            run_tests
            ;;
        lint)
            run_lint
            ;;
        dev)
            build_current true
            ;;
        release)
            build_release
            ;;
        docker)
            build_docker
            ;;
        help|*)
            show_help
            ;;
    esac
}

# 解析参数并运行
parse_args "$@"
main
