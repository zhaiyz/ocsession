.PHONY: build install clean test run build-all version help uninstall package-all

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS = -ldflags "-X github.com/zhaiyz/ocsession/internal/version.Version=$(VERSION) \
                    -X github.com/zhaiyz/ocsession/internal/version.GitCommit=$(GIT_COMMIT) \
                    -X github.com/zhaiyz/ocsession/internal/version.BuildDate=$(BUILD_DATE)"

build:
	CGO_ENABLED=1 go build $(LDFLAGS) -o bin/ocsession cmd/ocsession/main.go

install: build
	mkdir -p ~/.local/bin
	cp bin/ocsession ~/.local/bin/
	chmod +x ~/.local/bin/ocsession
	@echo ""
	@echo "✓ 已安装到 ~/.local/bin/ocsession"
	@echo ""
	@if ! echo "$$PATH" | grep -q "$$HOME/.local/bin"; then \
		echo "请将以下内容添加到 ~/.zshrc 或 ~/.bashrc:"; \
		echo "    export PATH=\"$$HOME/.local/bin:$$PATH\""; \
		echo ""; \
		echo "然后运行: source ~/.zshrc"; \
		echo ""; \
	fi
	@echo "运行: ocsession"

uninstall:
	rm -f ~/.local/bin/ocsession
	rm -f ~/.local/bin/ocsession.backup
	@echo "✓ 已卸载"

clean:
	rm -rf bin/
	rm -rf dist/
	go clean

test:
	go test -v ./test/unit/...

run:
	CGO_ENABLED=1 go run cmd/ocsession/main.go

build-all:
	@echo "构建所有平台..."
	@mkdir -p dist
	@echo "macos-arm64..."
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=1 go build $(LDFLAGS) -o dist/ocsession-macos-arm64 cmd/ocsession/main.go
	@echo "macos-amd64..."
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 go build $(LDFLAGS) -o dist/ocsession-macos-amd64 cmd/ocsession/main.go
	@echo "linux-amd64..."
	GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build $(LDFLAGS) -o dist/ocsession-linux-amd64 cmd/ocsession/main.go
	@echo "linux-arm64 (需要交叉编译工具链)..."
	GOOS=linux GOARCH=arm64 CC=aarch64-linux-gnu-gcc CGO_ENABLED=1 go build $(LDFLAGS) -o dist/ocsession-linux-arm64 cmd/ocsession/main.go
	@echo ""
	@echo "✓ 构建完成，文件在 dist/ 目录"

package-all: build-all
	@cd dist && \
	for f in ocsession-*; do \
		tar -czvf $$f.tar.gz $$f && \
		shasum -a 256 $$f.tar.gz > $$f.sha256; \
	done
	@echo "✓ 打包完成"

version:
	@echo "Version:    $(VERSION)"
	@echo "Commit:     $(GIT_COMMIT)"
	@echo "Build Date: $(BUILD_DATE)"
	@echo ""
	@if [ -f bin/ocsession ]; then \
		echo "Installed:"; \
		bin/ocsession -v; \
	fi

help:
	@echo "可用命令:"
	@echo "  make build        - 编译项目"
	@echo "  make install      - 安装到 ~/.local/bin"
	@echo "  make uninstall    - 卸载"
	@echo "  make clean        - 清理编译产物"
	@echo "  make test         - 运行测试"
	@echo "  make run          - 直接运行"
	@echo "  make build-all    - 构建所有平台"
	@echo "  make package-all  - 打包所有平台"
	@echo "  make version      - 显示版本信息"
	@echo ""
	@echo "环境变量:"
	@echo "  VERSION    - 自定义版本号"
	@echo "  GIT_COMMIT - 自定义 commit"
	@echo "  BUILD_DATE - 自定义构建时间"