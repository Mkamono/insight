package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"insight/src/models"
	"insight/src/usecase"

	"google.golang.org/genai"
	"gorm.io/gorm"
)

// DocumentGenerator はフラグメントからドキュメントを生成するサービス
type DocumentGenerator struct {
	client *genai.Client
	db     *gorm.DB
}

// NewDocumentGenerator は新しいDocumentGeneratorを作成
func NewDocumentGenerator(db *gorm.DB) (*DocumentGenerator, error) {
	ctx := context.Background()
	client, err := NewGenaiClient(ctx)
	if err != nil {
		return nil, err
	}

	return &DocumentGenerator{
		client: client,
		db:     db,
	}, nil
}

// DocumentRequest は単一ドキュメント作成要求を表す構造体
type DocumentRequest struct {
	Title       string   `json:"title"`
	Summary     string   `json:"summary"`
	Content     string   `json:"content"`
	FragmentIDs []int    `json:"fragment_ids"`
	Tags        []string `json:"tags"`
}

// DocumentsResponse は複数ドキュメント作成レスポンスを表す構造体
type DocumentsResponse struct {
	Documents []DocumentRequest `json:"documents"`
	Analysis  string            `json:"analysis"`
}

// GenerateDocuments はフラグメントからドキュメントを生成
func (g *DocumentGenerator) GenerateDocuments(ctx context.Context) error {
	fmt.Println("Fetching all fragments...")

	// 全フラグメントを取得
	fragmentUsecase := usecase.NewFragmentUsecase(g.db)
	fragments, err := fragmentUsecase.GetAllFragments()
	if err != nil {
		return fmt.Errorf("failed to get fragments: %w", err)
	}

	if len(fragments) == 0 {
		fmt.Println("No fragments found in database.")
		return nil
	}

	fmt.Printf("Found %d fragments. Analyzing with AI...\n", len(fragments))

	// フラグメントをIDでソートして順序を固定化
	for i := 0; i < len(fragments); i++ {
		for j := i + 1; j < len(fragments); j++ {
			if fragments[i].ID > fragments[j].ID {
				fragments[i], fragments[j] = fragments[j], fragments[i]
			}
		}
	}

	// AI生成実行
	response, err := g.generateDocumentsWithAI(ctx, fragments)
	if err != nil {
		return err
	}

	// ドキュメントを作成
	return g.createDocumentsFromResponse(response)
}

func (g *DocumentGenerator) generateDocumentsWithAI(ctx context.Context, fragments []models.Fragment) (*DocumentsResponse, error) {
	// フラグメント情報をプロンプトに構築
	fragmentsInfo := ""
	for _, fragment := range fragments {
		fragmentsInfo += fmt.Sprintf("Fragment ID: %d\nContent: %s\nCreated: %s\n\n",
			fragment.ID, fragment.Content, fragment.CreatedAt.Format("2006-01-02 15:04:05"))
	}

	// 構造化出力スキーマを定義
	schema := &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"documents": {
				Type: genai.TypeArray,
				Items: &genai.Schema{
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"title": {
							Type:        genai.TypeString,
							Description: "ドキュメントのタイトル",
						},
						"summary": {
							Type:        genai.TypeString,
							Description: "ドキュメントの要約",
						},
						"content": {
							Type:        genai.TypeString,
							Description: "ドキュメントの本文（Markdown形式）",
						},
						"fragment_ids": {
							Type: genai.TypeArray,
							Items: &genai.Schema{
								Type: genai.TypeInteger,
							},
							Description: "使用するフラグメントのIDリスト",
						},
						"tags": {
							Type: genai.TypeArray,
							Items: &genai.Schema{
								Type: genai.TypeString,
							},
							Description: "ドキュメントに適用するタグのリスト",
						},
					},
					Required: []string{"title", "summary", "content", "fragment_ids", "tags"},
				},
			},
			"analysis": {
				Type:        genai.TypeString,
				Description: "フラグメント分析の説明",
			},
		},
		Required: []string{"documents", "analysis"},
	}

	// 生成設定
	config := &genai.GenerateContentConfig{
		Temperature:      genai.Ptr(float32(0.1)), // より一貫した出力のために温度を下げる
		MaxOutputTokens:  8000,                    // 出力トークン数を増加
		ResponseMIMEType: "application/json",
		ResponseSchema:   schema,
	}

	// プロンプトを構築
	prompt := g.buildDocumentPrompt(fragmentsInfo)

	// AI生成を実行
	fmt.Println("Generating structured output...")
	resp, err := g.client.Models.GenerateContent(ctx, "gemini-2.5-flash", genai.Text(prompt), config)
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	// レスポンスをパース
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no response generated")
	}

	responseText := resp.Candidates[0].Content.Parts[0].Text

	var documentsResponse DocumentsResponse
	if err := json.Unmarshal([]byte(responseText), &documentsResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response JSON: %w", err)
	}

	// 分析結果を表示
	fmt.Printf("\n=== AI Analysis ===\n%s\n\n", documentsResponse.Analysis)

	return &documentsResponse, nil
}

func (g *DocumentGenerator) buildDocumentPrompt(fragmentsInfo string) string {
	return fmt.Sprintf(`以下のフラグメントを分析し、テーマ別にドキュメントを作成してください。

=== フラグメント一覧 ===
%s

=== 作成指針 ===
- 関連するフラグメントをテーマ別にグループ化
- フラグメントの内容を基に、必要な背景知識や詳細説明を適度に補完
- 実用的で読みやすいドキュメントを作成
- 各ドキュメントに適切なタグを2-4個程度付与

=== タグ付与基準 ===
- 短く簡潔で検索しやすいタグを使用

=== Markdown構造の要件 ===
contentフィールドは以下の構造に従ってください：
1. # メインタイトル（titleと同じ）
2. 空行
3. 導入段落（1-2行）
4. 空行
5. ## サブ見出し1
6. 内容段落（改行で区切る）
7. 空行
8. ## サブ見出し2
9. 内容段落
10. 必要に応じてさらなる見出しと内容

=== 出力形式 ===
JSON形式で、各ドキュメントには以下を含める：
- title: 適切な日本語タイトル
- summary: 1-2文の簡潔な要約
- content: 上記構造に従ったMarkdown形式の本文
- fragment_ids: 使用したフラグメントのIDリスト
- tags: ドキュメントに適したタグの配列

=== Markdown例 ===
# プログラミング言語の基礎

プログラミング言語は開発者がコンピュータに指示を与えるためのツールです。

## 主要な特徴

各言語には独自の特徴があります。パフォーマンス、開発効率、学習コストなどが選択の基準となります。

## 利用場面

Webアプリケーション開発、システム開発、機械学習など、用途に応じて適切な言語を選択することが重要です。`, fragmentsInfo)
}

func (g *DocumentGenerator) createDocumentsFromResponse(documentsResponse *DocumentsResponse) error {
	// ドキュメントを作成
	fmt.Printf("Creating %d documents...\n", len(documentsResponse.Documents))

	// 同一バッチのドキュメント群に同じバージョンタイムスタンプを設定
	versionCreatedAt := time.Now()
	fmt.Printf("Document version timestamp: %s\n\n", versionCreatedAt.Format("2006-01-02 15:04:05"))

	documentUsecase := usecase.NewDocumentUsecase(g.db)

	for i, docReq := range documentsResponse.Documents {
		fmt.Printf("Creating document %d: %s (Tags: %v)\n", i+1, docReq.Title, docReq.Tags)

		// FragmentIDsをuintスライスに変換
		fragmentIDs := make([]uint, len(docReq.FragmentIDs))
		for i, id := range docReq.FragmentIDs {
			fragmentIDs[i] = uint(id)
		}

		// タグを作成または取得
		tagIDs, err := g.createOrGetTags(docReq.Tags)
		if err != nil {
			fmt.Printf("Failed to create tags for document '%s': %v\n", docReq.Title, err)
			continue
		}

		input := usecase.CreateDocumentInput{
			Title:            docReq.Title,
			Summary:          docReq.Summary,
			Content:          docReq.Content,
			VersionCreatedAt: versionCreatedAt, // 同一バッチで同じタイムスタンプ
			FragmentIDs:      fragmentIDs,      // フラグメントとの関連付け
			TagIDs:           tagIDs,           // タグとの関連付け
		}

		document, err := documentUsecase.CreateDocument(input)
		if err != nil {
			fmt.Printf("Failed to create document '%s': %v\n", docReq.Title, err)
			continue
		}

		fmt.Printf("✓ Document created with ID: %d (Version: %s, Fragments: %d, Tags: %d)\n",
			document.ID,
			document.VersionCreatedAt.Format("2006-01-02 15:04:05"),
			len(fragmentIDs),
			len(tagIDs))
	}

	fmt.Println("\nDocument creation completed!")
	return nil
}

// createOrGetTags は指定されたタグ名のタグを作成または取得する
func (g *DocumentGenerator) createOrGetTags(tagNames []string) ([]uint, error) {
	var tagIDs []uint

	for _, tagName := range tagNames {
		if tagName == "" {
			continue
		}

		// 既存のタグを検索
		var tag models.Tag
		err := g.db.Where("name = ?", tagName).First(&tag).Error
		if err == nil {
			// 既存のタグが見つかった場合
			tagIDs = append(tagIDs, tag.ID)
			continue
		}

		// タグが見つからない場合は新しく作成
		newTag := models.Tag{
			Name:  tagName,
			Color: g.generateTagColor(tagName), // タグ名から色を生成
		}

		if err := g.db.Create(&newTag).Error; err != nil {
			return nil, fmt.Errorf("failed to create tag '%s': %w", tagName, err)
		}

		tagIDs = append(tagIDs, newTag.ID)
		fmt.Printf("  Created new tag: %s (ID: %d)\n", tagName, newTag.ID)
	}

	return tagIDs, nil
}

// generateTagColor はタグ名から一意の色を生成する
func (g *DocumentGenerator) generateTagColor(tagName string) string {
	// 簡単なハッシュベースの色生成
	hash := 0
	for _, char := range tagName {
		hash = int(char) + ((hash << 5) - hash)
	}

	// HSL色空間で彩度と明度を固定し、色相のみを変更
	hue := (hash%360 + 360) % 360

	// 色相を6つの主要な色に分類してより見やすい色にする
	colors := []string{"#3B82F6", "#EF4444", "#10B981", "#F59E0B", "#8B5CF6", "#EC4899"}
	return colors[hue%len(colors)]
}
