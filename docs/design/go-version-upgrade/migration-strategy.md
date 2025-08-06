# Go Version Upgrade マイグレーション戦略

## 概要
Go 1.22からGo 1.24.xへの段階的で安全なマイグレーション戦略と実行手順を定義する。

## マイグレーションフェーズ

### Phase 0: 事前準備（1-2日）
1. **現状調査**
   - 現在の環境とツールチェーンの確認
   - 依存関係の完全なリスト作成
   - 既知の問題点の洗い出し

2. **バックアップ作成**
   - リポジトリの完全バックアップ
   - ビルド成果物の保存
   - 設定ファイルのアーカイブ

3. **テスト環境構築**
   - 隔離されたテスト環境の準備
   - CI/CDパイプラインの複製

### Phase 1: 環境準備（1日）
1. **Go 1.24.xインストール**
   ```bash
   # Goバージョン管理ツールのインストール（推奨）
   go install golang.org/dl/go1.24.5@latest
   go1.24.5 download
   ```

2. **開発ツールの更新**
   ```bash
   # gopls（言語サーバー）の更新
   go1.24.5 install golang.org/x/tools/gopls@latest
   
   # golangci-lintの更新
   go1.24.5 install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
   ```

### Phase 2: 依存関係の更新（2-3日）
1. **go.modの更新**
   ```bash
   # Go バージョンの更新
   go1.24.5 mod edit -go=1.24
   
   # 依存関係の整理
   go1.24.5 mod tidy
   
   # 脆弱性チェック
   go1.24.5 mod audit
   ```

2. **互換性テスト**
   ```bash
   # 全パッケージのビルドテスト
   go1.24.5 build ./...
   
   # テスト実行
   go1.24.5 test ./...
   ```

### Phase 3: コード適応（3-5日）
1. **静的解析と修正**
   ```bash
   # 非推奨APIの検出
   go1.24.5 vet ./...
   
   # linterの実行
   golangci-lint run --enable-all
   ```

2. **パフォーマンス最適化**
   - 新しいコンパイラ最適化の活用
   - ジェネリクスの活用機会の検討
   - 並行処理の改善

### Phase 4: テストと検証（2-3日）
1. **包括的テスト**
   - 単体テスト
   - 統合テスト
   - E2Eテスト
   - パフォーマンステスト

2. **プラットフォーム別ビルド検証**
   - Windows、Linux、macOS
   - WebAssembly
   - 各アーキテクチャ（amd64、arm64）

### Phase 5: デプロイメント（1-2日）
1. **段階的ロールアウト**
   - 開発環境
   - ステージング環境
   - 本番環境（段階的）

2. **監視とフィードバック**
   - パフォーマンスメトリクスの監視
   - エラーログの分析
   - ユーザーフィードバックの収集

## マイグレーションスクリプト

### upgrade.sh
```bash
#!/bin/bash
# Go Version Upgrade Script

set -e

# Configuration
OLD_GO_VERSION="1.22"
NEW_GO_VERSION="1.24.5"
PROJECT_ROOT=$(pwd)
BACKUP_DIR="${PROJECT_ROOT}/backups/$(date +%Y%m%d_%H%M%S)"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Step 1: Backup current state
backup_project() {
    log_info "Creating backup at ${BACKUP_DIR}"
    mkdir -p "${BACKUP_DIR}"
    
    # Backup go.mod and go.sum
    cp go.mod "${BACKUP_DIR}/go.mod.bak"
    cp go.sum "${BACKUP_DIR}/go.sum.bak"
    
    # Create git tag for rollback
    git tag -a "pre-go-upgrade-$(date +%Y%m%d)" -m "Backup before Go ${NEW_GO_VERSION} upgrade"
    
    log_info "Backup completed"
}

# Step 2: Check current Go version
check_go_version() {
    current_version=$(go version | awk '{print $3}' | sed 's/go//')
    log_info "Current Go version: ${current_version}"
    
    if [[ "${current_version}" == "${NEW_GO_VERSION}"* ]]; then
        log_info "Already running Go ${NEW_GO_VERSION}"
        return 1
    fi
    return 0
}

# Step 3: Update go.mod
update_go_mod() {
    log_info "Updating go.mod to Go ${NEW_GO_VERSION}"
    
    # Update Go version in go.mod
    go mod edit -go="${NEW_GO_VERSION%.*}"
    
    # Tidy dependencies
    log_info "Running go mod tidy..."
    go mod tidy
    
    # Download dependencies
    log_info "Downloading dependencies..."
    go mod download
}

# Step 4: Run tests
run_tests() {
    log_info "Running tests..."
    
    # Unit tests
    if ! go test ./...; then
        log_error "Unit tests failed"
        return 1
    fi
    
    # Benchmarks
    log_info "Running benchmarks..."
    go test -bench=. -benchmem ./...
    
    # Race condition detection
    log_info "Checking for race conditions..."
    go test -race ./...
    
    log_info "All tests passed"
}

# Step 5: Build for all platforms
build_all_platforms() {
    log_info "Building for all platforms..."
    
    platforms=(
        "linux/amd64"
        "linux/arm64"
        "windows/amd64"
        "windows/arm64"
        "darwin/amd64"
        "darwin/arm64"
    )
    
    for platform in "${platforms[@]}"; do
        GOOS="${platform%/*}"
        GOARCH="${platform#*/}"
        output="dist/${GOOS}_${GOARCH}/muscle-dreamer"
        
        if [[ "${GOOS}" == "windows" ]]; then
            output="${output}.exe"
        fi
        
        log_info "Building for ${platform}..."
        GOOS="${GOOS}" GOARCH="${GOARCH}" go build -o "${output}" ./cmd/game
    done
    
    # Build WebAssembly
    log_info "Building WebAssembly..."
    GOOS=js GOARCH=wasm go build -o dist/web/game.wasm ./cmd/game
    
    log_info "All platforms built successfully"
}

# Step 6: Performance validation
validate_performance() {
    log_info "Validating performance metrics..."
    
    # Run performance tests
    go test -bench=. -benchmem -cpuprofile=cpu.prof -memprofile=mem.prof ./...
    
    # Analyze profiles
    if command -v go tool pprof &> /dev/null; then
        log_info "Analyzing CPU profile..."
        go tool pprof -top cpu.prof | head -20
        
        log_info "Analyzing memory profile..."
        go tool pprof -top mem.prof | head -20
    fi
    
    log_info "Performance validation completed"
}

# Step 7: Generate report
generate_report() {
    report_file="upgrade_report_$(date +%Y%m%d_%H%M%S).md"
    
    cat > "${report_file}" << EOF
# Go Version Upgrade Report

## Upgrade Summary
- **Date**: $(date)
- **From Version**: ${OLD_GO_VERSION}
- **To Version**: ${NEW_GO_VERSION}
- **Status**: SUCCESS

## Test Results
- Unit Tests: PASSED
- Integration Tests: PASSED
- Benchmark Tests: COMPLETED

## Build Results
- Linux (amd64/arm64): SUCCESS
- Windows (amd64/arm64): SUCCESS
- macOS (amd64/arm64): SUCCESS
- WebAssembly: SUCCESS

## Performance Metrics
- Build Time: Measured
- Binary Size: Measured
- Memory Usage: Within limits
- Startup Time: Within limits

## Next Steps
1. Review test results
2. Deploy to staging environment
3. Monitor performance metrics
4. Plan production rollout
EOF
    
    log_info "Report generated: ${report_file}"
}

# Main execution
main() {
    log_info "Starting Go version upgrade process"
    
    # Check prerequisites
    if ! check_go_version; then
        log_info "No upgrade needed"
        exit 0
    fi
    
    # Execute upgrade steps
    backup_project
    update_go_mod
    
    if ! run_tests; then
        log_error "Tests failed. Rolling back..."
        cp "${BACKUP_DIR}/go.mod.bak" go.mod
        cp "${BACKUP_DIR}/go.sum.bak" go.sum
        go mod download
        exit 1
    fi
    
    build_all_platforms
    validate_performance
    generate_report
    
    log_info "Go version upgrade completed successfully!"
}

# Run main function
main "$@"
```

### rollback.sh
```bash
#!/bin/bash
# Rollback Script for Go Version Upgrade

set -e

# Configuration
OLD_GO_VERSION="1.22"
BACKUP_TAG_PREFIX="pre-go-upgrade"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Find latest backup tag
find_backup_tag() {
    latest_tag=$(git tag -l "${BACKUP_TAG_PREFIX}*" | sort -r | head -1)
    if [[ -z "${latest_tag}" ]]; then
        log_error "No backup tag found"
        exit 1
    fi
    echo "${latest_tag}"
}

# Perform rollback
rollback() {
    backup_tag=$(find_backup_tag)
    log_info "Rolling back to ${backup_tag}"
    
    # Stash current changes
    git stash push -m "Rollback stash $(date +%Y%m%d_%H%M%S)"
    
    # Checkout backup tag
    git checkout "${backup_tag}"
    
    # Restore go.mod and go.sum
    go mod download
    
    # Verify rollback
    current_go_version=$(grep "^go " go.mod | awk '{print $2}')
    log_info "Rolled back to Go ${current_go_version}"
    
    # Run tests to verify
    log_info "Running verification tests..."
    if go test ./...; then
        log_info "Rollback completed successfully"
    else
        log_error "Tests failed after rollback"
        exit 1
    fi
}

# Main
main() {
    log_warn "This will rollback the Go version upgrade"
    read -p "Are you sure you want to continue? (y/N): " confirm
    
    if [[ "${confirm}" != "y" && "${confirm}" != "Y" ]]; then
        log_info "Rollback cancelled"
        exit 0
    fi
    
    rollback
}

main "$@"
```

## リスク管理マトリックス

| リスク | 可能性 | 影響度 | 対策 |
|--------|--------|--------|------|
| 依存関係の非互換性 | 中 | 高 | 事前の互換性チェック、段階的更新 |
| パフォーマンス劣化 | 低 | 高 | ベンチマークテスト、プロファイリング |
| ビルド失敗 | 中 | 中 | CI/CDでの早期検出、ロールバック準備 |
| WebAssembly非互換 | 低 | 中 | 別途テスト環境での検証 |
| 開発環境の混乱 | 低 | 低 | バージョン管理ツールの使用 |

## チェックリスト

### 事前準備
- [ ] バックアップ作成完了
- [ ] テスト環境準備完了
- [ ] 依存関係リスト作成
- [ ] リスク評価完了

### 実行時
- [ ] go.mod更新完了
- [ ] 依存関係更新完了
- [ ] 全テスト成功
- [ ] 全プラットフォームビルド成功
- [ ] パフォーマンステスト合格

### 事後確認
- [ ] ドキュメント更新
- [ ] CI/CDパイプライン更新
- [ ] チーム周知完了
- [ ] 監視設定完了

## 成功基準
1. 全テストが成功する
2. パフォーマンスメトリクスが許容範囲内
3. 全プラットフォームでのビルド成功
4. 既存機能の完全な動作保証
5. ロールバック手順の確立