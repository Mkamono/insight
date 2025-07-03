# Insight - ドキュメント品質向上システム

フラグメントの曖昧性を自動検出し、質問生成によってドキュメント品質を向上させるAI駆動型知識管理システムです。

## 特徴

- **曖昧性の自動検出**: 人物名の表記ゆれ、時期の不明確さ、専門用語の定義不足を自動検出
- **質問駆動型改善**: AIが生成する質問に回答することで段階的にドキュメント品質を向上
- **時系列考慮処理**: 新しい情報を優先し、情報の更新や変更を適切に反映
- **Web対応設計**: 全機能がCLI実装済み、REST API化が容易

## セットアップ

### 1. 環境変数の設定

`.env.example`をコピーして`.env`ファイルを作成し、APIキーを設定してください：

```bash
cp .env.example .env
```

`.env`ファイルを編集：
```bash
# Gemini API Key for AI processing
GEMINI_API_KEY=your_actual_gemini_api_key_here

# Database file path (optional)
INSIGHT_DB_FILE=knowledge.db
```

### 2. ビルド

```bash
go build -o insight ./cmd/insight
```

## 基本的な使い方

### 1. フラグメント追加
```bash
./insight add "プロジェクトAの進捗について田中さんと打ち合わせした"
./insight add "最近、新機能の開発を開始した"
```

### 2. AI処理（質問生成+ドキュメント作成）
```bash
./insight process --ai
```

出力例：
```
Generated question (ID 1): 「田中さん」とは具体的にどなたですか？
Generated question (ID 2): 「最近」の新機能開発はいつ開始しましたか？
Generated question (ID 3): 「新機能」とは具体的にどのような機能ですか？
New document 'プロジェクト進捗管理' created successfully with ID 1.
```

### 3. 質問確認と回答
```bash
# 質問一覧表示
./insight question list

# 質問に回答
./insight answer 1 "田中さんは開発チームのリーダーです"
./insight answer 2 "新機能開発は2025年7月1日に開始しました"
```

### 4. 再処理で品質向上
```bash
./insight process --ai
```

## コマンド一覧

### 基本操作
- `./insight add "テキスト"` - フラグメント追加
- `./insight process --ai` - AI処理実行
- `./insight status` - システム状況確認

### 質問・回答システム
- `./insight question list` - 質問一覧表示
- `./insight question list --pending` - 未回答質問のみ表示
- `./insight answer <ID> "回答"` - 質問に回答

### ドキュメント管理
- `./insight doc list` - ドキュメント一覧
- `./insight doc show <ID>` - ドキュメント詳細表示
- `./insight find --tag "タグ名"` - タグ検索
- `./insight find --query "検索語"` - 全文検索

### システム管理
- `./insight clear --confirm` - 全データクリア
- `./insight reset --confirm` - ドキュメントのみリセット
- `./insight export` - Markdown形式でエクスポート
- `./insight web --port 8080` - Webサーバー起動

## アーキテクチャ

```
insight/
├── cmd/insight/main.go          # エントリーポイント
├── internal/                    # 機能別パッケージ
│   ├── ai/processor.go          # AI処理
│   ├── cli/app.go              # CLI処理
│   ├── database/database.go     # データアクセス層
│   ├── export/export.go         # エクスポート機能
│   └── web/server.go           # Webサーバー
├── sql/                        # データベーススキーマ
├── e2e/                        # E2Eテスト
├── .env.example                # 環境変数テンプレート
└── .gitignore                  # Git無視ファイル
```

## 質問生成の仕組み

AIは以下の曖昧性を自動検出して質問を生成します：

1. **人物・組織の同定**: 「田中さん」「斎藤さん」などの表記ゆれ
2. **時期・日付**: 「最近」「先日」「今度」などの不明確な時間表現
3. **場所・部署**: 「あの部署」「向こうのチーム」などの指示代名詞
4. **専門用語・略語**: 「例の件」「あのシステム」などの具体性不足
5. **数値・仕様**: 「多くの」「大幅な」などの曖昧な数量表現

## テスト

```bash
# E2Eテスト実行
cd e2e && go test -v

# ユニットテスト実行
go test ./...
```

## ライセンス

MIT License
