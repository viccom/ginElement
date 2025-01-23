# 项目名称
PROJECT_NAME := goiot

# Go 编译器
GO := go

# 支持的平台和架构
PLATFORMS := linux
ARCHS := 386 amd64 arm arm64

# 编译输出目录
OUTPUT_DIR := release

# swag init
swag:
	@echo "Generating Swagger documentation..."
	@swag init

# 默认目标：编译所有平台和架构
all: clean $(PLATFORMS)

# 编译指定平台的所有架构
$(PLATFORMS):
	@$(foreach ARCH, $(ARCHS), \
		$(MAKE) build PLATFORM=$@ ARCH=$(ARCH); \
	)

# 编译单个平台和架构
build:
	@echo "Building for $(PLATFORM)/$(ARCH)..."
	@mkdir -p $(OUTPUT_DIR)/$(PLATFORM)/$(ARCH)	
	@sed '/OpcDARead/ s/^/\/\//' handler/handler.go
	@CGO_ENABLED=0 GOOS=$(PLATFORM) GOARCH=$(ARCH) $(GO) build -ldflags="-s -w" -o $(OUTPUT_DIR)/$(PLATFORM)/$(ARCH)/$(PROJECT_NAME)$(if $(filter windows, $(PLATFORM)),.exe,) .
	@sed -i '/OpcDARead/ s/^\/\///' handler/handler.go
	@echo "Compressing binary with upx..."
	@upx --best $(OUTPUT_DIR)/$(PLATFORM)/$(ARCH)/$(PROJECT_NAME)$(if $(filter windows, $(PLATFORM)),.exe,)

# 清理编译输出
clean:
	@echo "Cleaning build directory..."
	@rm -rf $(OUTPUT_DIR)

# 帮助信息
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  all       Build for all platforms and architectures (default)"
	@echo "  clean     Clean the build directory"
	@echo "  build     Build for a specific platform and architecture (requires PLATFORM and ARCH)"
	@echo "  help      Show this help message"
	@echo ""
	@echo "Examples:"
	@echo "  make all"
	@echo "  make build PLATFORM=linux ARCH=amd64"
	@echo "  make clean"

.PHONY: all $(PLATFORMS) build clean help
