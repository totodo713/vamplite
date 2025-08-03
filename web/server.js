const express = require('express');
const compression = require('compression');
const cors = require('cors');
const path = require('path');

const app = express();
const PORT = process.env.PORT || 3000;

// ミドルウェア
app.use(compression());
app.use(cors());

// WebAssembly MIMEタイプ設定
app.use((req, res, next) => {
    if (req.path.endsWith('.wasm')) {
        res.type('application/wasm');
    }
    next();
});

// 静的ファイル配信
app.use(express.static('.'));

// メインページ
app.get('/', (req, res) => {
    res.sendFile(path.join(__dirname, 'index.html'));
});

// ヘルスチェック
app.get('/health', (req, res) => {
    res.json({ 
        status: 'OK', 
        timestamp: new Date().toISOString(),
        version: process.env.npm_package_version || '0.1.0'
    });
});

app.listen(PORT, () => {
    console.log(`🌐 Web開発サーバーが起動しました: http://localhost:${PORT}`);
    console.log(`📊 ヘルスチェック: http://localhost:${PORT}/health`);
});
