# Third-Party Licenses

このドキュメントは、Muscle Dreamerプロジェクトで使用されているサードパーティライブラリのライセンス情報をまとめたものです。

## 直接依存関係

### Ebitengine v2.6.3
- **Repository**: https://github.com/hajimehoshi/ebiten
- **License**: Apache License 2.0
- **Description**: 2Dゲーム開発エンジン
- **Usage**: メインのゲームエンジンとして使用

## 間接依存関係

### Go標準ライブラリ拡張 (golang.org/x/*)

#### golang.org/x/image
- **License**: BSD 3-Clause License
- **Description**: 画像処理ライブラリ
- **Usage**: 画像フォーマットの読み書き

#### golang.org/x/mobile
- **License**: BSD 3-Clause License  
- **Description**: モバイルプラットフォーム対応
- **Usage**: Android/iOSビルド対応

#### golang.org/x/sync
- **License**: BSD 3-Clause License
- **Description**: 同期処理ユーティリティ
- **Usage**: 並行処理制御

#### golang.org/x/sys
- **License**: BSD 3-Clause License
- **Description**: システムコール
- **Usage**: OS固有の低レベル操作

#### golang.org/x/exp/shiny
- **License**: BSD 3-Clause License
- **Description**: 実験的UIライブラリ
- **Usage**: ウィンドウ管理・入力処理

#### golang.org/x/text
- **License**: BSD 3-Clause License
- **Description**: テキスト処理
- **Usage**: 文字エンコーディング

#### golang.org/x/tools
- **License**: BSD 3-Clause License
- **Description**: Go開発ツール
- **Usage**: ビルド時ツール

#### golang.org/x/mod
- **License**: BSD 3-Clause License
- **Description**: Goモジュール管理
- **Usage**: モジュール依存関係解決

### Ebitengineエコシステム

#### github.com/ebitengine/purego v0.5.0
- **License**: Apache License 2.0
- **Description**: Pure Goでのネイティブライブラリ呼び出し
- **Usage**: プラットフォーム固有機能の呼び出し

#### github.com/ebitengine/oto/v3 v3.1.0
- **License**: Apache License 2.0
- **Description**: 低レベルオーディオライブラリ
- **Usage**: サウンド出力

### 音声処理ライブラリ

#### github.com/hajimehoshi/go-mp3 v0.3.4
- **License**: Apache License 2.0
- **Description**: MP3デコーダー
- **Usage**: MP3音声ファイルの読み込み

#### github.com/jfreymuth/oggvorbis v1.0.5
- **License**: MIT License
- **Description**: Ogg Vorbisデコーダー
- **Usage**: OGG音声ファイルの読み込み

#### github.com/jfreymuth/vorbis v1.0.2
- **License**: MIT License
- **Description**: Vorbis音声コーデック
- **Usage**: Vorbis音声デコード

### グラフィックス・UI関連

#### github.com/go-gl/glfw/v3.3/glfw
- **License**: zlib/libpng License
- **Description**: OpenGLコンテキスト管理
- **Usage**: ウィンドウ作成・入力処理

#### github.com/go-text/typesetting
- **License**: MIT License (推定)
- **Description**: テキストレンダリング
- **Usage**: フォント描画・テキスト配置

#### github.com/hajimehoshi/bitmapfont/v3 v3.0.0
- **License**: Apache License 2.0
- **Description**: ビットマップフォント
- **Usage**: テキスト描画

### システム・その他

#### github.com/jezek/xgb v1.1.0
- **License**: BSD 3-Clause License
- **Description**: X11プロトコルバインディング
- **Usage**: LinuxでのX11連携

#### github.com/jakecoffman/cp v1.2.1
- **License**: MIT License
- **Description**: 物理エンジン（Chipmunk Physics）
- **Usage**: 物理シミュレーション

#### dmitri.shuralyov.com/gpu/mtl
- **License**: MIT License
- **Description**: Metal API バインディング
- **Usage**: macOS GPU アクセス

## ライセンス互換性の評価

### 本プロジェクト（CC BY-NC 4.0）との互換性

✅ **互換性あり:**
- Apache License 2.0: 商用利用制限との競合なし
- MIT License: 制限が少なく互換
- BSD 3-Clause License: 制限が少なく互換
- zlib/libpng License: 制限が少なく互換

### 注意事項

1. **CC BY-NC 4.0の制約**: 本プロジェクト全体は非商用利用に制限されています
2. **帰属表示**: Apache/MIT/BSDライセンスは帰属表示が必要です
3. **配布時の義務**: 各ライセンス条項に従った適切な表示が必要です

## 配布時に含める必要があるライセンス文書

配布パッケージには以下のライセンス文書を含める必要があります：

1. 本プロジェクトのLICENSE（CC BY-NC 4.0）
2. 各サードパーティライブラリのライセンス文書
3. このTHIRD_PARTY_LICENSES.mdファイル

## 更新日

- 作成日: 2025年8月9日
- 最終更新日: 2025年8月9日

---

**注意**: このドキュメントは依存関係の変更に応じて定期的に更新する必要があります。新しいライブラリを追加する際は、必ずライセンス互換性を確認してください。
