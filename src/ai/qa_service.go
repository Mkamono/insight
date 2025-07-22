package ai

import (
	"context"
	"fmt"
	"insight/src/models"
	"time"

	"google.golang.org/genai"
	"gorm.io/gorm"
)

// QAService はドキュメントに対する質問応答機能を提供するサービス
type QAService struct {
	client *genai.Client
	db     *gorm.DB
}

// NewQAService は新しいQAServiceを作成
func NewQAService(db *gorm.DB) (*QAService, error) {
	ctx := context.Background()
	client, err := NewGenaiClient(ctx)
	if err != nil {
		return nil, err
	}

	return &QAService{
		client: client,
		db:     db,
	}, nil
}

// QARequest は質問応答リクエストを表す構造体
type QARequest struct {
	DocumentID   uint   `json:"document_id" validate:"required"`
	Question     string `json:"question" validate:"required"`
	UseWebSearch bool   `json:"use_web_search"`
}

// GlobalQARequest は全ドキュメント対象の質問応答リクエストを表す構造体
type GlobalQARequest struct {
	Question     string `json:"question" validate:"required"`
	UseWebSearch bool   `json:"use_web_search"`
}

// QAResponse は質問応答レスポンスを表す構造体
type QAResponse struct {
	Answer    string   `json:"answer"`
	Sources   []string `json:"sources"`
	WebSearch bool     `json:"web_search_used"`
}

// AskQuestion はドキュメントに対する質問に回答する
func (s *QAService) AskQuestion(ctx context.Context, req QARequest) (*QAResponse, error) {
	// ドキュメントを取得
	var document models.Document
	if err := s.db.Preload("Fragments").Preload("Tags").First(&document, req.DocumentID).Error; err != nil {
		return nil, fmt.Errorf("failed to find document: %w", err)
	}

	// コンテキスト構築
	context := s.buildContext(&document)

	// AIに質問（Web検索は内部で自動的に処理される）
	response, err := s.generateAnswer(ctx, req.Question, context, "", req.UseWebSearch)
	if err != nil {
		return nil, fmt.Errorf("failed to generate answer: %w", err)
	}

	// 情報源を構築
	sources := []string{fmt.Sprintf("Document: %s", document.Title)}
	if req.UseWebSearch {
		sources = append(sources, "Web search results (when available)")
	}

	return &QAResponse{
		Answer:    response,
		Sources:   sources,
		WebSearch: req.UseWebSearch,
	}, nil
}

// buildContext はドキュメントからコンテキストを構築
func (s *QAService) buildContext(document *models.Document) string {
	context := fmt.Sprintf(`=== Document Information ===
Title: %s
Summary: %s

=== Content ===
%s

=== Related Fragments ===
`, document.Title, document.Summary, document.Content)

	// 関連フラグメントを追加
	for i, fragment := range document.Fragments {
		context += fmt.Sprintf("Fragment %d: %s\n", i+1, fragment.Content)
	}

	// タグ情報を追加
	if len(document.Tags) > 0 {
		context += "\n=== Tags ===\n"
		for _, tag := range document.Tags {
			context += fmt.Sprintf("- %s\n", tag.Name)
		}
	}

	return context
}

// min はint型の最小値を返すヘルパー関数
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// generateAnswer はAIを使用して回答を生成
func (s *QAService) generateAnswer(ctx context.Context, question, documentContext, webContext string, useWebSearch bool) (string, error) {
	// プロンプト構築
	prompt := s.buildAnswerPrompt(question, documentContext, webContext, useWebSearch)
	
	// デバッグ: プロンプトの最初の500文字を表示
	fmt.Printf("Generated prompt (first 500 chars): %s...\n", prompt[:min(len(prompt), 500)])

	// 生成設定
	config := &genai.GenerateContentConfig{
		Temperature:     genai.Ptr(float32(0.3)),
		MaxOutputTokens: 2000,
	}

	// Web検索を有効にする場合はGoogle Search toolを設定
	if useWebSearch {
		config.Tools = []*genai.Tool{{
			GoogleSearch: &genai.GoogleSearch{},
		}}
	}

	// AI生成実行
	resp, err := s.client.Models.GenerateContent(ctx, "gemini-2.0-flash-exp", genai.Text(prompt), config)
	if err != nil {
		return "", fmt.Errorf("failed to generate response: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no response generated")
	}

	return resp.Candidates[0].Content.Parts[0].Text, nil
}

// buildAnswerPrompt は回答生成用のプロンプトを構築
func (s *QAService) buildAnswerPrompt(question, documentContext, webContext string, useWebSearch bool) string {
	prompt := fmt.Sprintf(`あなたは技術文書の専門家です。提供されたドキュメントの内容に基づいて、ユーザーの質問に正確で有用な回答を提供してください。

=== ドキュメント情報 ===
%s

=== 質問 ===
%s

=== 回答指針 ===
- ドキュメントの内容を第一に参考にしてください
- ドキュメントに記載されていない情報については、一般的な知識で補完しても構いませんが、その旨を明記してください`, documentContext, question)

	// Web検索が有効な場合は追加の指針を含める
	if useWebSearch {
		prompt += `
- 必要に応じてWeb検索を使用して最新の情報を取得してください
- Web検索結果を使用した場合は、その旨を明記してください
- ドキュメント内容とWeb検索結果を組み合わせて、より包括的な回答を提供してください`
	}

	prompt += `
- 回答は具体的で実用的なものにしてください
- 必要に応じてコード例や手順を含めてください
- 日本語で回答してください`

	return prompt
}

// AskGlobalQuestion は最新バージョンのドキュメントを対象とした質問に回答する
func (s *QAService) AskGlobalQuestion(ctx context.Context, req GlobalQARequest) (*QAResponse, error) {
	// 最新バージョンを取得
	var latestVersion time.Time
	if err := s.db.Model(&models.Document{}).Select("version_created_at").Order("version_created_at DESC").Limit(1).Scan(&latestVersion).Error; err != nil {
		return nil, fmt.Errorf("failed to get latest version: %w", err)
	}

	if latestVersion.IsZero() {
		return &QAResponse{
			Answer:    "現在、ドキュメントが存在しません。",
			Sources:   []string{"No documents found"},
			WebSearch: req.UseWebSearch,
		}, nil
	}

	// 最新バージョンのドキュメントのみを取得
	var documents []models.Document
	if err := s.db.Preload("Fragments").Preload("Tags").Where("version_created_at = ?", latestVersion).Find(&documents).Error; err != nil {
		return nil, fmt.Errorf("failed to find documents: %w", err)
	}

	fmt.Printf("Global QA: Found %d documents in latest version (%s)\n", len(documents), latestVersion.Format("2006-01-02 15:04:05"))
	for i, doc := range documents {
		fmt.Printf("Document %d: ID=%d, Title=%s, Fragments=%d, Tags=%d\n", 
			i+1, doc.ID, doc.Title, len(doc.Fragments), len(doc.Tags))
	}

	if len(documents) == 0 {
		return &QAResponse{
			Answer:    "現在、最新バージョンにドキュメントが存在しません。",
			Sources:   []string{"No documents found in latest version"},
			WebSearch: req.UseWebSearch,
		}, nil
	}

	// 全ドキュメントのコンテキストを構築
	context := s.buildGlobalContext(documents)
	fmt.Printf("Global context length: %d characters\n", len(context))

	// AIに質問
	response, err := s.generateAnswer(ctx, req.Question, context, "", req.UseWebSearch)
	if err != nil {
		return nil, fmt.Errorf("failed to generate answer: %w", err)
	}

	// 情報源を構築
	sources := []string{fmt.Sprintf("Latest version documents (%d total)", len(documents))}
	if req.UseWebSearch {
		sources = append(sources, "Web search results (when available)")
	}

	return &QAResponse{
		Answer:    response,
		Sources:   sources,
		WebSearch: req.UseWebSearch,
	}, nil
}

// buildGlobalContext は最新バージョンのドキュメントからコンテキストを構築
func (s *QAService) buildGlobalContext(documents []models.Document) string {
	context := fmt.Sprintf("=== Latest Version Document Collection ===\nDocuments in latest version: %d\n\n", len(documents))
	fmt.Printf("Building global context for %d latest version documents\n", len(documents))

	// 各ドキュメントの情報を追加（最大10ドキュメントまで詳細表示）
	maxDocs := 10
	for i, document := range documents {
		if i >= maxDocs {
			context += fmt.Sprintf("... and %d more documents\n", len(documents)-maxDocs)
			break
		}

		context += fmt.Sprintf("=== Document %d: %s ===\n", i+1, document.Title)
		context += fmt.Sprintf("Summary: %s\n", document.Summary)
		
		// タグ情報を追加
		if len(document.Tags) > 0 {
			context += "Tags: "
			for j, tag := range document.Tags {
				if j > 0 {
					context += ", "
				}
				context += tag.Name
			}
			context += "\n"
		}

		// コンテンツの要約版を追加（長すぎる場合は切り取り）
		content := document.Content
		if len(content) > 500 {
			content = content[:500] + "..."
		}
		context += fmt.Sprintf("Content: %s\n\n", content)
	}

	// 全体的なタグ統計を追加
	tagCounts := make(map[string]int)
	for _, document := range documents {
		for _, tag := range document.Tags {
			tagCounts[tag.Name]++
		}
	}

	if len(tagCounts) > 0 {
		context += "=== Available Tags ===\n"
		for tag, count := range tagCounts {
			context += fmt.Sprintf("- %s (%d documents)\n", tag, count)
		}
		context += "\n"
	}

	return context
}
