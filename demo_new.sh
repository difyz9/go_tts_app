#!/bin/bash

# TTS应用快速演示脚本 - 全新版本
# 展示主要功能和特色

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 项目信息
APP_NAME="tts_app"

# 函数定义
print_header() {
    echo -e "${BLUE}=================================${NC}"
    echo -e "${BLUE}🎵 TTS语音合成应用 - 功能演示${NC}"
    echo -e "${BLUE}=================================${NC}"
    echo ""
}

print_step() {
    echo -e "${CYAN}📍 $1${NC}"
    echo ""
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_info() {
    echo -e "${YELLOW}💡 $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# 检查应用是否存在
check_app() {
    if [ ! -f "./$APP_NAME" ]; then
        print_error "未找到 $APP_NAME 可执行文件"
        echo "请先构建应用："
        echo "  make build"
        echo "  # 或"
        echo "  go build -o $APP_NAME"
        exit 1
    fi
}

# 创建演示输入文件
create_demo_files() {
    print_step "创建演示文件"
    
    # 中文演示文本
    cat > demo_chinese.txt << EOF
欢迎使用TTS语音合成应用
这是一个功能强大的文本转语音工具
支持腾讯云TTS和Microsoft Edge TTS两种引擎
Edge TTS完全免费，无需API密钥
并发处理模式可以显著提高转换速度
智能文本过滤会自动跳过空行和无效内容
支持多种音色和语言选择
实时进度显示让您了解处理状态
一键音频合并功能自动整合所有片段
跨平台支持，可在Windows、macOS、Linux上运行
EOF

    # 英文演示文本
    cat > demo_english.txt << EOF
Welcome to TTS Voice Synthesis Application
This is a powerful text-to-speech tool
Supports Tencent Cloud TTS and Microsoft Edge TTS
Edge TTS is completely free without API keys
Concurrent processing mode significantly improves conversion speed
Intelligent text filtering automatically skips empty and invalid content
Supports multiple voice types and language selections
Real-time progress display keeps you informed of processing status
One-click audio merging function automatically integrates all fragments
Cross-platform support for Windows, macOS, and Linux
EOF

    print_success "演示文件创建完成"
    echo "  📄 demo_chinese.txt - 中文演示"
    echo "  📄 demo_english.txt - 英文演示"
    echo ""
}

# 显示应用信息
show_app_info() {
    print_step "应用信息"
    
    ./$APP_NAME --version
    echo ""
    
    print_info "支持的命令："
    echo "  🎯 edge  - Edge TTS (免费)"
    echo "  🏢 tts   - 腾讯云TTS (需API密钥)"
    echo "  🔗 merge - 音频合并"
    echo ""
}

# 演示Edge TTS功能
demo_edge_tts() {
    print_step "演示 Edge TTS (免费引擎)"
    
    print_info "查看可用的中文语音..."
    ./$APP_NAME edge --list zh
    echo ""
    
    print_info "使用默认语音转换中文文本..."
    ./$APP_NAME edge -i demo_chinese.txt -o output_chinese/
    
    if [ $? -eq 0 ]; then
        print_success "中文转换完成！输出目录: output_chinese/"
    fi
    echo ""
}

# 显示结果统计
show_results() {
    print_step "处理结果统计"
    
    echo "📊 输出目录统计："
    
    for dir in output_*; do
        if [ -d "$dir" ]; then
            file_count=$(find "$dir" -name "*.mp3" | wc -l)
            if [ -f "$dir/merged_audio.mp3" ]; then
                file_size=$(ls -lh "$dir/merged_audio.mp3" | awk '{print $5}')
                echo "  📁 $dir - $file_count 个文件，合并文件大小: $file_size"
            else
                echo "  📁 $dir - $file_count 个文件"
            fi
        fi
    done
    echo ""
    
    print_info "可以使用以下命令播放生成的音频："
    echo "  # macOS:"
    echo "  afplay output_chinese/merged_audio.mp3"
    echo ""
    echo "  # Linux:"
    echo "  mpv output_chinese/merged_audio.mp3"
    echo ""
}

# 清理功能
cleanup() {
    print_step "清理演示文件"
    
    read -p "是否删除演示生成的文件？(y/N): " -n 1 -r
    echo ""
    
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        rm -rf output_*
        rm -f demo_*.txt
        print_success "清理完成"
    else
        print_info "保留演示文件，您可以继续查看和播放生成的音频"
    fi
}

# 显示帮助信息
show_help() {
    print_header
    echo "TTS应用演示脚本"
    echo ""
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  demo       运行演示（默认）"
    echo "  info       只显示应用信息"
    echo "  clean      清理演示文件"
    echo "  help       显示此帮助信息"
    echo ""
}

# 主函数
main() {
    case ${1:-"demo"} in
        demo)
            print_header
            check_app
            create_demo_files
            show_app_info
            demo_edge_tts
            show_results
            cleanup
            ;;
        info)
            print_header
            check_app
            show_app_info
            ;;
        clean)
            cleanup
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            print_error "未知选项: $1"
            show_help
            exit 1
            ;;
    esac
}

# 运行主函数
main "$@"
