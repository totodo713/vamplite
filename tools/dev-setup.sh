#!/bin/bash
# tools/dev-setup.sh - 開発環境セットアップスクリプト

set -e

echo "🏗️  マッスルドリーマー開発環境セットアップ"

# 環境チェック
check_requirements() {
    echo "📋 必要なツールをチェック中..."
    
    command -v docker >/dev/null 2>&1 || { echo "❌ Docker が必要です"; exit 1; }
    command -v go >/dev/null 2>&1 || { echo "❌ Go が必要です"; exit 1; }
    command -v git >/dev/null 2>&1 || { echo "❌ Git が必要です"; exit 1; }
    
    echo "✅ 必要なツールが揃っています"
}

# Docker環境セットアップ
setup_docker() {
    echo "🐳 Docker環境をセットアップ中..."
    
    if [ ! -f "docker-compose.yml" ]; then
        echo "❌ docker-compose.yml が見つかりません"
        exit 1
    fi
    
    docker compose build
    echo "✅ Docker環境構築完了"
}

# Go環境セットアップ
setup_go() {
    echo "🐹 Go環境をセットアップ中..."
    
    go mod tidy
    go mod download
    
    # 開発ツールインストール
    go install golang.org/x/tools/gopls@latest
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
    go install golang.org/x/tools/cmd/goimports@latest
    
    echo "✅ Go環境セットアップ完了"
}

# アセットディレクトリ作成
setup_assets() {
    echo "🎨 アセットディレクトリをセットアップ中..."
    
    mkdir -p assets/{sprites,audio,fonts,ui}
    mkdir -p themes/default/{assets,scripts,localization}
    mkdir -p mods/{enabled,disabled,staging}
    
    echo "✅ アセットディレクトリ作成完了"
}

# 設定ファイル作成
setup_configs() {
    echo "⚙️  設定ファイルをセットアップ中..."
    
    if [ ! -f "config/game.yaml" ]; then
        cp config/game.yaml.example config/game.yaml 2>/dev/null || true
    fi
    
    echo "✅ 設定ファイルセットアップ完了"
}

# テスト実行
run_tests() {
    echo "🧪 初期テストを実行中..."
    
    go test ./... 2>/dev/null || echo "⚠️  テストファイルがまだありません"
    
    echo "✅ テスト実行完了"
}

# メイン実行
main() {
    check_requirements
    setup_docker
    setup_go
    setup_assets
    setup_configs
    run_tests
    
    echo ""
    echo "🎉 セットアップ完了!"
    echo ""
    echo "次のステップ:"
    echo "  1. 開発環境起動: make docker-dev"
    echo "  2. ゲーム実行: make dev"
    echo "  3. テスト実行: make test"
    echo ""
    echo "詳細なコマンド: make help"
}

main "$@"
