# マッスルドリーマー Web環境

WebAssembly版マッスルドリーマーの開発環境です。

## セットアップ

```bash
cd web
npm install
```

## 開発

```bash
# 開発サーバー起動
npm run dev

# WebAssemblyビルド（Goプロジェクトルートから）
make build-web
```

## ビルド

```bash
# 本番ビルド
npm run build

# 本番サーバー起動
npm start
```

## ディレクトリ構造

```
web/
├── index.html          # メインHTMLページ
├── server.js          # Express開発サーバー
├── package.json       # Node.js依存関係
├── game.wasm          # Goから生成されるWebAssemblyファイル
├── wasm_exec.js       # Go WebAssemblyランタイム
├── assets/           # ゲームアセット
├── src/              # WebフロントエンドJSコード
└── dist/             # ビルド成果物
```
