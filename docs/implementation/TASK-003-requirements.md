# TASK-003: 基本コンポーネント実装 - 要件定義

## 概要

ECSフレームワークの基本コンポーネント5種類を実装します。これらは2Dゲームで最も頻繁に使用される基本的なデータコンテナです。

## 実装対象コンポーネント

### 1. TransformComponent
**目的**: エンティティの位置、回転、スケールを管理

**要件**:
- Position (Vector2): 2D座標位置
- Rotation (float64): Z軸回転角度（ラジアン）
- Scale (Vector2): X,Y軸スケール値
- Parent/Child階層関係のサポート
- ワールド座標とローカル座標の変換
- 高速な行列計算

**制約**:
- メモリ使用量: 40バイト以下
- 行列計算: O(1)時間複雑度
- スレッドセーフ

### 2. SpriteComponent  
**目的**: 2D画像表示のための描画情報

**要件**:
- TextureID (string): テクスチャリソース識別子
- SourceRect (AABB): スプライトシート内の切り出し領域
- Color (Color): 色調整・透明度
- ZOrder (int): 描画順序
- Visible (bool): 表示/非表示フラグ
- FlipX/FlipY (bool): 反転フラグ

**制約**:
- メモリ使用量: 64バイト以下
- 描画時のソート効率性を考慮
- テクスチャリソースの参照管理

### 3. PhysicsComponent
**目的**: 物理シミュレーション用のパラメータ

**要件**:
- Velocity (Vector2): 速度ベクトル
- Acceleration (Vector2): 加速度ベクトル
- Mass (float64): 質量
- Friction (float64): 摩擦係数
- Gravity (bool): 重力の影響を受けるか
- IsStatic (bool): 静的オブジェクトフラグ
- MaxSpeed (float64): 最大速度制限

**制約**:
- メモリ使用量: 56バイト以下  
- 数値計算の精度と安定性
- 物理エンジンとの統合性

### 4. HealthComponent
**目的**: エンティティの体力・状態管理

**要件**:
- CurrentHealth (int): 現在の体力値
- MaxHealth (int): 最大体力値
- Shield (int): シールド値
- IsInvincible (bool): 無敵状態フラグ
- LastDamageTime (time.Time): 最後のダメージ時刻
- RegenerationRate (float64): 体力回復レート
- StatusEffects ([]StatusEffect): 状態異常リスト

**制約**:
- メモリ使用量: 72バイト以下
- 状態変更イベントの発火
- 並行アクセス時の整合性

### 5. AIComponent
**目的**: NPCの人工知能制御

**要件**:
- State (AIState): 現在のAI状態
- Target (EntityID): ターゲットエンティティ
- PatrolPoints ([]Vector2): 巡回ポイント
- DetectionRadius (float64): 探知範囲
- AttackRange (float64): 攻撃範囲
- Speed (float64): 移動速度
- Behavior (AIBehavior): 行動パターン
- LastStateChange (time.Time): 状態変更時刻

**制約**:
- メモリ使用量: 96バイト以下
- 状態機械の効率的な実装
- リアルタイム意思決定

## 共通要件

### パフォーマンス要件
- コンポーネント作成: < 1μs
- データアクセス: < 100ns
- シリアライゼーション: < 10μs
- メモリアライメント: 8バイト境界

### 機能要件
- Component interfaceの完全実装
- シリアライゼーション/デシリアライゼーション
- ディープコピー機能
- データ検証機能
- デバッグ表示機能

### 品質要件
- 単体テストカバレッジ: > 95%
- ベンチマークテスト: 性能要件の検証
- メモリリーク検出: valgrind/race detector
- 並行アクセステスト: race condition検出

### 技術制約
- Go言語標準の型安全性
- GCによるメモリ管理最適化
- reflect パッケージの使用は最小限
- unsafe パッケージの使用禁止

## 受け入れ基準

### 機能面
- [ ] 全5コンポーネントが Component interface を実装
- [ ] シリアライゼーションが正常に動作
- [ ] データ検証が適切に実行される
- [ ] 階層関係の管理が正常に動作（Transform）
- [ ] 状態機械が効率的に動作（AI）

### 性能面
- [ ] メモリ使用量が制約内に収まる
- [ ] アクセス時間が要件を満たす
- [ ] 大量のコンポーネント（10,000個）で安定動作
- [ ] GCへの影響が最小限

### 品質面
- [ ] 単体テストが全て成功
- [ ] ベンチマークテストが性能要件をクリア
- [ ] レースコンディションが検出されない
- [ ] メモリリークが発生しない

## 実装順序

1. **基底構造の定義**: 共通の構造体と定数
2. **TransformComponent**: 最も基本的なコンポーネント
3. **SpriteComponent**: 描画系の基盤
4. **PhysicsComponent**: 物理シミュレーション基盤
5. **HealthComponent**: ゲーム状態管理
6. **AIComponent**: 最も複雑なコンポーネント

## リスク要因

### 技術リスク
- 数値計算の精度問題（Physics）
- メモリアライメントの最適化
- 並行アクセス時の競合状態

### 設計リスク
- Interface設計の柔軟性
- 将来の拡張性確保
- 他システムとの統合性

### パフォーマンスリスク
- GCの影響によるフレームドロップ
- メモリ断片化
- キャッシュミスの増大

## 成功指標

- 全てのコンポーネントが要件を満たす
- 10,000エンティティで60FPS維持
- メモリ使用量 < 100MB（10,000エンティティ）
- 単体テスト成功率 100%
- ベンチマーク要件達成率 100%