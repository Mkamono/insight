# Insight

AI駆動型知識ドキュメント化・リンクシステム

## 概要

Insightは、情報の断片（フラグメント）を自動的に整理・分析し、AIを活用して構造化されたドキュメントを生成するシステムです。散らばった情報を効率的に管理し、知識として体系化することで、より良い意思決定をサポートします。

## 主な機能

### 📝 フラグメント管理
- テキスト、URL、画像の情報断片を収集・保存
- 階層的な親子関係でフラグメントを整理
- 処理状況の追跡とバッチ処理

### 🤖 AI駆動ドキュメント生成
- 複数フラグメントから意味のあるドキュメントを自動生成
- Markdown形式での構造化されたコンテンツ作成
- 自動要約とタグ付け機能

### 🔗 インテリジェントリンク
- フラグメントとドキュメントの自動関連付け
- タグベースのカテゴリ化とフィルタリング
- 関連性の高い情報の発見

### 🖥️ 多様なインターフェース
- **CLI**: コマンドライン操作で高速な情報処理
- **Web UI**: 直感的なブラウザベースの管理画面
- **API**: 外部システムとの連携

### 📁 自動ファイル生成
- Markdownファイルの自動生成と管理
- `/knowledge/documents/`ディレクトリでの整理
- メタデータテーブルでの構造化情報表示

## アーキテクチャ

```
insight/
├── core/                 # コア機能ライブラリ
│   ├── db/              # データベース接続・スキーマ
│   ├── services/        # ビジネスロジック
│   └── types/           # TypeScript型定義
├── cli/                 # コマンドラインインターフェース
├── web/                 # Next.js Webアプリケーション
├── knowledge/           # 生成されたファイル
│   ├── data.db         # SQLiteデータベース
│   └── documents/      # 生成されたMarkdownファイル
└── dist/               # コンパイル済みJavaScript
```

## インストール・セットアップ

### 必要環境
- Node.js 18以上
- [mise](https://mise.jdx.dev/) （推奨）
- Google Gemini API キー

### 1. リポジトリのクローン
```bash
git clone <repository-url>
cd insight
```

### 2. 依存関係のインストール
```bash
# mise使用の場合（推奨）
mise install

# 手動インストール
npm install
cd web && pnpm install
```

### 3. 環境変数の設定
```bash
# .env.local を作成
cp .env.example .env.local
```

`.env.local`を編集してAPIキーを設定：
```
GEMINI_API_KEY=your_gemini_api_key_here
```

### 4. データベース初期化
```bash
npm run build
node dist/cli/index.js init
```

## 基本的な使用方法

### CLIでの操作

#### フラグメントの作成
```bash
# テキストフラグメントの作成
node dist/cli/index.js fragment create -c "Docker はコンテナ化技術で、アプリケーションを軽量で移植可能な環境で実行できます"

# URLありフラグメントの作成
node dist/cli/index.js fragment create -c "Kubernetes公式ドキュメント" -u "https://kubernetes.io/docs/"
```

#### AI処理の実行
```bash
# 未処理フラグメントをバッチ処理
node dist/cli/index.js ai process-all
```

#### データの確認
```bash
# フラグメント一覧
node dist/cli/index.js fragment list

# ドキュメント一覧
node dist/cli/index.js document list
```

### Webインターフェース

```bash
# 開発サーバーの起動
cd web
pnpm run dev
```

ブラウザで `http://localhost:9342` にアクセス

## 技術スタック

### バックエンド
- **TypeScript**: タイプセーフな開発
- **SQLite**: 軽量データベース
- **Drizzle ORM**: TypeScript-first ORM
- **Google Gemini API**: AI機能

### フロントエンド
- **Next.js 15**: React フレームワーク
- **Tailwind CSS**: スタイリング
- **react-markdown**: Markdownレンダリング

### 開発ツール
- **mise**: 開発環境管理
- **ESLint**: コード品質
- **Commander.js**: CLI構築

## データベース設計

### 主要テーブル
- **fragments**: 情報断片（テキスト、URL、画像）
- **documents**: AI生成ドキュメント
- **tags**: カテゴリ分類用タグ
- **questions**: システム質問

### 関係
- Fragment ↔ Document: 多対多
- Document ↔ Tag: 多対多
- Fragment: 階層構造（親子関係）

## 開発コマンド

```bash
# TypeScriptビルド
npm run build

# 開発サーバー起動
cd web && pnpm run dev

# 型チェック
npm run typecheck

# リント
npm run lint
```

## 生成されるファイル

AIによって生成されるMarkdownファイルは以下の構造を持ちます：

```markdown
## 概要
AIが生成したコンテンツ...

---

## メタデータ

| 項目               | 内容               |
| ------------------ | ------------------ |
| **要約**           | ドキュメントの要約 |
| **タグ**           | タグ1, タグ2       |
| **作成日**         | 2025-01-01         |
| **更新日**         | 2025-01-01         |
| **ドキュメントID** | 1                  |

## 参考フラグメント

| ID  | 内容               | URL | 画像     |
| --- | ------------------ | --- | -------- |
| 1   | フラグメントの内容 | URL | 画像パス |
```

## AIの活用方法

### バッチ処理
複数のフラグメントを一度に処理し、APIコール数を削減：
- 関連性の高いフラグメントをグループ化
- 効率的なドキュメント生成
- レート制限の回避

### 自動化機能
- フラグメントからドキュメントへの自動変換
- タグの自動生成と分類
- 関連情報の自動リンク

## 今後の計画

- [ ] 画像解析機能の強化
- [ ] 質問応答システムの実装
- [ ] 検索機能の拡充
- [ ] エクスポート機能の追加
- [ ] 他システムとの連携機能

## ライセンス

MIT License

## 作成者

[あなたの名前]
