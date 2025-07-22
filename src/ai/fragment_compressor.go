package ai

import (
	"context"
	"encoding/json"
	"fmt"

	"insight/src/models"
	"insight/src/usecase"

	"google.golang.org/genai"
	"gorm.io/gorm"
)

// FragmentCompressor はフラグメントの圧縮を行うサービス
type FragmentCompressor struct {
	client *genai.Client
	db     *gorm.DB
}

// NewFragmentCompressor は新しいFragmentCompressorを作成
func NewFragmentCompressor(db *gorm.DB) (*FragmentCompressor, error) {
	ctx := context.Background()
	client, err := NewGenaiClient(ctx)
	if err != nil {
		return nil, err
	}

	return &FragmentCompressor{
		client: client,
		db:     db,
	}, nil
}

// CompressFragments はフラグメントを圧縮する
func (c *FragmentCompressor) CompressFragments(ctx context.Context) error {
	fmt.Println("Fetching all fragments for compression...")

	// 全フラグメントを取得
	fragmentUsecase := usecase.NewFragmentUsecase(c.db)
	fragments, err := fragmentUsecase.GetAllFragments()
	if err != nil {
		return fmt.Errorf("failed to get fragments: %w", err)
	}

	if len(fragments) < 2 {
		fmt.Println("Need at least 2 fragments for compression.")
		return nil
	}

	fmt.Printf("Found %d fragments. Analyzing for compression...\n", len(fragments))

	// AI分析を実行
	response, err := c.analyzeFragmentsWithAI(ctx, fragments)
	if err != nil {
		return err
	}

	// アクションを実行
	return c.executeCompressionActions(response)
}

func (c *FragmentCompressor) analyzeFragmentsWithAI(ctx context.Context, fragments []models.Fragment) (*compressionResponse, error) {
	// フラグメント情報をプロンプトに構築
	fragmentsInfo := ""
	for _, fragment := range fragments {
		fragmentsInfo += fmt.Sprintf("Fragment ID: %d\nContent: %s\n\n",
			fragment.ID, fragment.Content)
	}

	// 構造化出力スキーマを定義
	schema := &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"actions": {
				Type: genai.TypeArray,
				Items: &genai.Schema{
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"type": {
							Type:        genai.TypeString,
							Description: "アクション種別: 'merge' または 'delete'",
						},
						"fragment_ids": {
							Type: genai.TypeArray,
							Items: &genai.Schema{
								Type: genai.TypeInteger,
							},
							Description: "対象フラグメントのIDリスト",
						},
						"new_content": {
							Type:        genai.TypeString,
							Description: "統合時の新しい内容（mergeの場合のみ）",
						},
						"reason": {
							Type:        genai.TypeString,
							Description: "アクションの理由",
						},
					},
					Required: []string{"type", "fragment_ids", "reason"},
				},
			},
			"summary": {
				Type:        genai.TypeString,
				Description: "圧縮処理の要約",
			},
		},
		Required: []string{"actions", "summary"},
	}

	// 生成設定
	config := &genai.GenerateContentConfig{
		Temperature:      genai.Ptr(float32(0.1)),
		ResponseMIMEType: "application/json",
		ResponseSchema:   schema,
	}

	// プロンプトを構築
	prompt := c.buildCompressionPrompt(fragmentsInfo)

	// AI分析を実行
	fmt.Println("Analyzing fragments for compression...")
	resp, err := c.client.Models.GenerateContent(ctx, "gemini-2.5-flash", genai.Text(prompt), config)
	if err != nil {
		return nil, fmt.Errorf("failed to generate compression analysis: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no compression analysis generated")
	}

	responseText := resp.Candidates[0].Content.Parts[0].Text

	// レスポンスをパース
	var compressionResponse compressionResponse
	if err := json.Unmarshal([]byte(responseText), &compressionResponse); err != nil {
		return nil, fmt.Errorf("failed to parse compression response: %w", err)
	}

	fmt.Printf("\n=== Compression Analysis ===\n%s\n\n", compressionResponse.Summary)

	return &compressionResponse, nil
}

type compressionResponse struct {
	Actions []struct {
		Type        string `json:"type"`
		FragmentIDs []int  `json:"fragment_ids"`
		NewContent  string `json:"new_content,omitempty"`
		Reason      string `json:"reason"`
	} `json:"actions"`
	Summary string `json:"summary"`
}

func (c *FragmentCompressor) buildCompressionPrompt(fragmentsInfo string) string {
	return fmt.Sprintf(`以下のフラグメントを分析し、圧縮のためのアクションを提案してください。

=== フラグメント一覧 ===
%s

=== 圧縮基準 ===
1. **類似内容の統合**: 内容が重複または類似するフラグメントを統合
2. **情報量の少ないフラグメントの削除**: 意味のない、情報量の極めて少ないフラグメントを削除
3. **品質向上**: より具体的で有用な内容に統合

=== アクション ===
- **merge**: 複数のフラグメントを統合し、new_contentで新しい内容を作成
- **delete**: 情報量が少ない、または不要なフラグメントを削除

=== 制約 ===
- 削除しすぎないこと（最大でも全体の30%%程度まで）
- 統合時は元の情報を失わないこと
- 重要な情報は必ず保持すること

JSON形式で出力してください。`, fragmentsInfo)
}

func (c *FragmentCompressor) executeCompressionActions(response *compressionResponse) error {
	// アクションを実行
	for i, action := range response.Actions {
		fmt.Printf("Action %d: %s (IDs: %v) - %s\n", i+1, action.Type, action.FragmentIDs, action.Reason)

		if action.Type == "merge" && len(action.FragmentIDs) > 1 {
			// 統合処理
			err := c.mergeFragments(action.FragmentIDs, action.NewContent)
			if err != nil {
				fmt.Printf("Failed to merge fragments %v: %v\n", action.FragmentIDs, err)
				continue
			}
			fmt.Printf("✓ Merged fragments %v\n", action.FragmentIDs)

		} else if action.Type == "delete" && len(action.FragmentIDs) > 0 {
			// 削除処理
			err := c.deleteFragments(action.FragmentIDs)
			if err != nil {
				fmt.Printf("Failed to delete fragments %v: %v\n", action.FragmentIDs, err)
				continue
			}
			fmt.Printf("✓ Deleted fragments %v\n", action.FragmentIDs)
		}
	}

	fmt.Println("\nFragment compression completed!")
	return nil
}

// mergeFragments は複数のフラグメントを統合する
func (c *FragmentCompressor) mergeFragments(fragmentIDs []int, newContent string) error {
	// 最初のフラグメントを更新
	firstID := uint(fragmentIDs[0])

	// トランザクション内で処理
	return c.db.Transaction(func(tx *gorm.DB) error {
		// 最初のフラグメントを更新
		if err := tx.Model(&models.Fragment{}).Where("id = ?", firstID).Update("content", newContent).Error; err != nil {
			return err
		}

		// 残りのフラグメントを削除
		for i := 1; i < len(fragmentIDs); i++ {
			if err := tx.Delete(&models.Fragment{}, fragmentIDs[i]).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// deleteFragments は指定されたフラグメントを削除する
func (c *FragmentCompressor) deleteFragments(fragmentIDs []int) error {
	return c.db.Transaction(func(tx *gorm.DB) error {
		for _, id := range fragmentIDs {
			if err := tx.Delete(&models.Fragment{}, id).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
