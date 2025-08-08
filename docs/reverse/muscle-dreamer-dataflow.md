# Muscle Dreamer データフロー設計（逆生成）

## 分析概要

**分析日時**: 2025-08-07  
**対象コードベース**: /home/devman/GolandProjects/muscle-dreamer  
**分析対象**: Go 1.24 + Ebitengine v2.6.3 + ECS フレームワーク  
**信頼度**: 92%（実装済みコードとテストに基づく分析）

## データフロー概要

### メインゲームループ

```mermaid
sequenceDiagram
    participant M as Main
    participant G as Game
    participant W as World
    participant SM as SystemManager
    participant EM as EntityManager
    participant CM as ComponentManager
    participant E as Ebitengine

    M->>G: NewGame()
    G->>W: Initialize World
    W->>SM: Initialize Systems
    W->>EM: Initialize EntityManager
    W->>CM: Initialize ComponentManager
    
    loop Game Loop (60 FPS)
        E->>G: Update()
        G->>W: Update(deltaTime)
        W->>SM: UpdateSystems(deltaTime)
        SM->>SM: ExecuteSystemsByPriority()
        
        E->>G: Draw(screen)
        G->>W: Render(screen)
        W->>SM: RenderSystems(screen)
    end
    M->>G: core.NewGame()
    G-->>M: *Game インスタンス
    M->>E: ebiten.RunGame(game)
    
    loop ゲームループ
        E->>G: Update()
        G-->>E: nil (成功)
        E->>G: Draw(screen)
        G->>C: 背景色描画
        G->>C: デバッグテキスト描画
        G-->>E: 描画完了
        E->>G: Layout(w, h)
        G-->>E: (1280, 720)
    end
    
    Note over E: 60FPS target
```

### 2. WebAssembly起動フロー

```mermaid
sequenceDiagram
    participant U as ユーザー
    participant B as ブラウザ
    participant H as index.html
    participant W as WebAssembly
    participant G as Go Runtime
    participant E as Ebitengine
    participant C as HTML5 Canvas
    
    U->>B: Webページアクセス
    B->>H: HTML読み込み
    H->>B: wasm_exec.js読み込み
    H->>W: game.wasm フェッチ
    W-->>H: WebAssembly モジュール
    H->>G: new Go() + instantiate
    G->>E: ゲーム初期化
    
    loop Web ゲームループ
        E->>G: Update()
        E->>C: Canvas描画
        C->>B: ブラウザレンダリング
        B->>U: 画面表示
    end
```

## ゲームエンジン内部データフロー

### 3. Ebitengine ゲームループ

```mermaid
flowchart TD
    A[ゲーム開始] --> B[初期化]
    B --> C{ゲームループ}
    
    C --> D[Update 呼び出し]
    D --> E[ゲーム状態更新]
    E --> F[Draw 呼び出し]
    F --> G[画面描画]
    G --> H[Layout 呼び出し]
    H --> I[画面サイズ計算]
    I --> J{終了条件?}
    
    J -->|継続| C
    J -->|終了| K[ゲーム終了]
    
    style C fill:#e1f5fe
    style J fill:#fff3e0
```

### 4. 現在実装されている描画フロー

```mermaid
flowchart LR
    A[Draw メソッド呼び出し] --> B[背景色設定]
    B --> C[screen.Fill RGBA 50,50,100,255]
    C --> D[デバッグテキスト描画]
    D --> E[ebitenutil.DebugPrint]
    E --> F[描画完了]
    
    style A fill:#c8e6c9
    style F fill:#c8e6c9
```

## 設定・データ管理フロー

### 5. 設定ファイル読み込みフロー (将来実装)

```mermaid
flowchart TD
    A[アプリケーション開始] --> B[config/game.yaml 読み込み]
    B --> C{ファイル存在?}
    C -->|Yes| D[YAML パース]
    C -->|No| E[デフォルト設定使用]
    
    D --> F[設定値検証]
    F --> G{設定値有効?}
    G -->|Yes| H[Game構造体に設定適用]
    G -->|No| I[エラーログ + デフォルト値]
    
    E --> H
    I --> H
    H --> J[ゲーム初期化完了]
    
    style B fill:#e3f2fd
    style H fill:#e8f5e8
```

### 6. アセット管理フロー (将来実装)

```mermaid
flowchart TD
    A[アセット読み込み要求] --> B{アセットタイプ}
    
    B -->|スプライト| C[assets/sprites/]
    B -->|オーディオ| D[assets/audio/]
    B -->|UI| E[assets/ui/]
    B -->|フォント| F[assets/fonts/]
    
    C --> G[画像ファイル読み込み]
    D --> H[オーディオファイル読み込み]
    E --> I[UI要素読み込み]
    F --> J[フォントファイル読み込み]
    
    G --> K[メモリキャッシュ]
    H --> K
    I --> K
    J --> K
    
    K --> L[ゲームで使用可能]
    
    style K fill:#fff9c4
```

## テーマ・MODシステムデータフロー (設計済み)

### 7. テーマシステムフロー

```mermaid
flowchart TD
    A[テーマ選択] --> B[themes/{theme-name}/]
    B --> C[theme.yaml 読み込み]
    C --> D[テーマメタデータ解析]
    D --> E[テーマアセット読み込み]
    
    E --> F[themes/{theme}/assets/]
    E --> G[themes/{theme}/localization/]
    E --> H[themes/{theme}/scripts/]
    
    F --> I[アセット置換]
    G --> J[言語リソース置換]
    H --> K[スクリプト実行]
    
    I --> L[ゲーム適用]
    J --> L
    K --> L
    
    style B fill:#e1f5fe
    style L fill:#e8f5e8
```

### 8. MODシステムフロー (セキュア設計)

```mermaid
flowchart TD
    A[MOD読み込み] --> B[mods/enabled/]
    B --> C[MODメタデータ検証]
    C --> D{セキュリティチェック}
    
    D -->|Pass| E[サンドボックス環境作成]
    D -->|Fail| F[MOD無効化]
    
    E --> G[制限されたファイルアクセス]
    E --> H[制限されたAPI]
    E --> I[ネットワークアクセス禁止]
    
    G --> J[MODスクリプト実行]
    H --> J
    I --> J
    
    J --> K[ゲーム機能拡張]
    F --> L[エラーログ出力]
    
    style E fill:#fff3e0
    style K fill:#e8f5e8
    style F fill:#ffebee
```

## Webアーキテクチャデータフロー

### 9. Web開発サーバーフロー

```mermaid
sequenceDiagram
    participant B as ブラウザ
    participant E as Express Server
    participant F as ファイルシステム
    participant W as WebAssembly
    
    B->>E: GET /
    E->>F: index.html 読み込み
    F-->>E: HTMLファイル
    E-->>B: HTML レスポンス
    
    B->>E: GET /game.wasm
    E->>F: game.wasm 読み込み
    Note over E: MIMEタイプ: application/wasm
    F-->>E: WebAssembly バイナリ
    E-->>B: WASM レスポンス
    
    B->>E: GET /wasm_exec.js
    E->>F: Go WebAssembly ランタイム
    F-->>E: JavaScript ファイル
    E-->>B: JS レスポンス
    
    B->>E: GET /health
    E-->>B: {"status": "OK", "timestamp": "..."}
```

### 10. クロスプラットフォームビルドフロー

```mermaid
flowchart TD
    A[ソースコード] --> B[Make コマンド]
    B --> C{ビルドターゲット}
    
    C -->|build| D[ローカルビルド]
    C -->|build-web| E[WebAssembly ビルド]
    C -->|build-all| F[全プラットフォーム]
    C -->|docker-build| G[Docker ビルド]
    
    D --> H[dist/muscle-dreamer]
    E --> I[dist/web/game.wasm]
    E --> J[dist/web/wasm_exec.js]
    
    F --> K[Windows バイナリ]
    F --> L[Linux バイナリ]
    F --> M[macOS バイナリ]
    F --> I
    
    G --> N[Docker コンテナ内ビルド]
    N --> O[クロスコンパイル成果物]
    
    style C fill:#e3f2fd
    style H fill:#e8f5e8
    style I fill:#e8f5e8
```

## エラーハンドリング・ログフロー

### 11. エラー処理フロー (現在の実装)

```mermaid
flowchart TD
    A[エラー発生] --> B{エラー種別}
    
    B -->|ゲーム初期化エラー| C[log.Fatal err]
    B -->|Update エラー| D[nil return 継続]
    B -->|描画エラー| E[panic recovery]
    B -->|WebAssembly エラー| F[JavaScript console.error]
    
    C --> G[プロセス終了]
    D --> H[ゲーム継続]
    E --> H
    F --> I[ブラウザエラー表示]
    
    style G fill:#ffebee
    style H fill:#e8f5e8
    style I fill:#fff3e0
```

### 12. 将来のログシステムフロー

```mermaid
flowchart TD
    A[ログイベント] --> B[ログレベル判定]
    B --> C{レベル}
    
    C -->|DEBUG| D[デバッグ情報]
    C -->|INFO| E[一般情報]
    C -->|WARN| F[警告]
    C -->|ERROR| G[エラー]
    C -->|FATAL| H[致命的エラー]
    
    D --> I[stdout 出力]
    E --> I
    F --> J[stderr 出力]
    G --> J
    H --> K[ログファイル + 終了]
    
    I --> L[開発時表示]
    J --> L
    K --> M[エラー解析]
    
    style K fill:#ffebee
    style L fill:#e8f5e8
```

## パフォーマンス・監視データフロー

### 13. ゲームパフォーマンス監視 (将来実装)

```mermaid
flowchart TD
    A[フレーム開始] --> B[Update 実行時間測定]
    B --> C[Draw 実行時間測定]
    C --> D[メモリ使用量取得]
    D --> E[FPS 計算]
    
    E --> F{パフォーマンス基準}
    F -->|60FPS達成| G[正常動作]
    F -->|60FPS未満| H[パフォーマンス警告]
    
    G --> I[統計情報更新]
    H --> J[最適化提案]
    H --> I
    
    I --> K[デバッグ表示 F3]
    J --> K
    
    style G fill:#e8f5e8
    style H fill:#fff3e0
    style K fill:#e3f2fd
```

## セーブ・設定データフロー

### 14. ゲームセーブフロー (将来実装)

```mermaid
sequenceDiagram
    participant P as プレイヤー
    participant G as ゲーム
    participant S as セーブシステム
    participant F as ファイルシステム
    
    P->>G: セーブ操作
    G->>S: セーブデータ作成
    S->>S: ゲーム状態シリアライズ
    S->>F: saves/save_*.json 書き込み
    F-->>S: 書き込み完了
    S-->>G: セーブ成功
    G-->>P: セーブ完了通知
    
    Note over S: JSON形式で状態保存
    Note over F: ローカルファイルシステム
```

### 15. 設定変更フロー (将来実装)

```mermaid
flowchart TD
    A[設定変更要求] --> B[設定UI]
    B --> C[バリデーション]
    C --> D{設定値有効?}
    
    D -->|Yes| E[設定適用]
    D -->|No| F[エラーメッセージ]
    
    E --> G[config/game.yaml 更新]
    E --> H[ゲーム設定リロード]
    
    G --> I[設定永続化]
    H --> J[即座に反映]
    
    F --> K[設定修正要求]
    K --> B
    
    style E fill:#e8f5e8
    style F fill:#ffebee
    style J fill:#e3f2fd
```

## 開発・ビルド環境データフロー

### 16. Docker開発環境フロー

```mermaid
flowchart TD
    A[開発者] --> B[docker compose up]
    B --> C{サービス選択}
    
    C -->|dev| D[メイン開発環境]
    C -->|web-dev| E[Web開発環境]
    C -->|cross-compile| F[クロスコンパイル環境]
    C -->|test| G[テスト環境]
    
    D --> H[Go開発 + デバッガー]
    E --> I[Node.js + WebAssembly]
    F --> J[マルチプラットフォームビルド]
    G --> K[テスト実行]
    
    H --> L[ローカル開発]
    I --> M[Web開発サーバー]
    J --> N[配布用バイナリ]
    K --> O[品質保証]
    
    style L fill:#e8f5e8
    style M fill:#e3f2fd
    style N fill:#fff9c4
    style O fill:#e1f5fe
```

## データフロー分析まとめ

### 現在実装済みフロー
✅ **基本ゲームループ**: Update → Draw → Layout  
✅ **WebAssembly起動**: HTML → WASM → Canvas  
✅ **クロスプラットフォームビルド**: Go → 複数バイナリ  
✅ **Docker開発環境**: 複数サービス並行開発  

### 設計済み・未実装フロー
🔄 **設定管理**: YAML → ゲーム設定  
🔄 **アセット管理**: ファイル → メモリキャッシュ  
🔄 **テーマシステム**: テーマフォルダ → ゲーム適用  
🔄 **MODシステム**: セキュアサンドボックス実行  

### 将来実装予定フロー
⏳ **セーブシステム**: ゲーム状態 → JSON永続化  
⏳ **エラーハンドリング**: 構造化ログ → ファイル出力  
⏳ **パフォーマンス監視**: リアルタイム → 統計表示  
⏳ **設定UI**: 動的設定変更 → 即座反映  

### データフロー設計評価
- **シンプルさ**: ⭐⭐⭐⭐⭐ 明確で理解しやすいフロー
- **拡張性**: ⭐⭐⭐⭐⭐ プラグイン・テーマ対応設計
- **パフォーマンス**: ⭐⭐⭐⭐☆ 60FPS目標、最適化の余地
- **セキュリティ**: ⭐⭐⭐⭐☆ MODサンドボックス設計
- **保守性**: ⭐⭐⭐⭐☆ Go言語の恩恵、構造化設計

### 推奨実装順序
1. **設定管理システム** - アプリケーション動作の基盤
2. **アセット管理システム** - ゲームコンテンツ読み込み
3. **基本的なECSフレームワーク** - ゲームエンティティ管理
4. **エラーハンドリング・ログシステム** - デバッグ・運用基盤
5. **パフォーマンス監視** - 品質保証機能