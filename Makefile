# マッスルドリーマー開発用 Makefile

.PHONY: help dev build test clean docker-setup docker-dev docker-build

# デフォルトターゲット
help:
	@echo "マッスルドリーマー開発コマンド"
	@echo ""
	@echo "開発環境:"
	@echo "  docker-setup    - Docker環境初期化"
	@echo "  docker-dev      - 開発環境起動"
	@echo "  dev            - ローカル開発サーバー起動"
	@echo ""
	@echo "ビルド:"
	@echo "  build          - デバッグビルド"
	@echo "  build-release  - リリースビルド"
	@echo "  build-web      - WebAssemblyビルド"
	@echo "  build-all      - 全プラットフォームビルド"
	@echo ""
	@echo "テスト:"
	@echo "  test           - ユニットテスト実行"
	@echo "  test-integration - 統合テスト実行"
	@echo "  test-all       - 全テスト実行"
	@echo ""
	@echo "ツール:"
	@echo "  lint           - コード解析"
	@echo "  format         - コードフォーマット"
	@echo "  clean          - ビルド成果物削除"

# Docker環境セットアップ
docker-setup:
	@echo "Docker環境をセットアップ中..."
	docker compose build
	docker compose run --rm dev go mod tidy
	@echo "セットアップ完了!"

# Docker開発環境起動
docker-dev:
	docker compose up -d dev web-dev
	@echo "開発環境が起動しました:"
	@echo "  - メイン開発: http://localhost:8080"
	@echo "  - Web開発: http://localhost:3000"
	@echo ""
	@echo "開発コンテナに接続: docker compose exec dev bash"

# ローカル開発
dev:
	go run cmd/game/main.go

# ビルドターゲット
build:
	mkdir -p dist
	go build -o dist/muscle-dreamer cmd/game/main.go

build-release:
	mkdir -p dist
	go build -ldflags="-s -w" -o dist/muscle-dreamer cmd/game/main.go

build-web:
	mkdir -p dist/web
	GOOS=js GOARCH=wasm go build -o dist/web/game.wasm cmd/game/main.go
	cp "$$(go env GOROOT)/misc/wasm/wasm_exec.js" dist/web/

build-all:
	mkdir -p dist/{windows,linux,darwin,web}
	
	# Windows
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o dist/windows/muscle-dreamer.exe cmd/game/main.go
	
	# Linux
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o dist/linux/muscle-dreamer cmd/game/main.go
	
	# macOS
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o dist/darwin/muscle-dreamer cmd/game/main.go
	GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o dist/darwin/muscle-dreamer-arm64 cmd/game/main.go
	
	# WebAssembly
	GOOS=js GOARCH=wasm go build -o dist/web/game.wasm cmd/game/main.go
	cp "$$(go env GOROOT)/misc/wasm/wasm_exec.js" dist/web/

# Docker内でのクロスコンパイル
docker-build:
	docker compose run --rm cross-compile

# テスト
test:
	go test ./...

test-integration:
	go test -tags=integration ./tests/...

test-all:
	go test -v ./...
	go test -tags=integration -v ./tests/...

# コード品質
lint:
	golangci-lint run --skip-dirs docs

format:
	go fmt ./...
	goimports -w .

# クリーンアップ
clean:
	rm -rf dist/
	docker compose down -v
	docker system prune -f

# 依存関係更新
deps:
	go mod tidy
	go mod download

# プロジェクト初期化
init:
	go mod init muscle-dreamer
	@echo "module muscle-dreamer" > go.mod
	@echo "" >> go.mod
	@echo "go 1.22" >> go.mod
	@echo "" >> go.mod
	@echo "require (" >> go.mod
	@echo "    github.com/hajimehoshi/ebiten/v2 v2.6.0" >> go.mod
	@echo "    gopkg.in/yaml.v3 v3.0.1" >> go.mod
	@echo ")" >> go.mod
	go mod tidy
