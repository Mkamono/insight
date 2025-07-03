package ai

import (
	"context"
	"database/sql"
	"fmt"
	"insight/internal/database"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

func getGeminiAPIKey() string {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Fatal("GEMINI_API_KEY environment variable is required")
	}
	return apiKey
}

// GeneratedDocument は生成されたドキュメントの構造体です
type GeneratedDocument struct {
	Title   string
	Content string
	Summary string
}

// ProcessingDecision は処理の判断結果を表す構造体です
type ProcessingDecision struct {
	Action           string // "create_new", "append_to_existing", "merge_documents"
	TargetDocumentID int    // 対象ドキュメントのID（該当する場合）
	Reasoning        string // 判断理由
}

// ProcessFragments は未処理のフラグメントをAIで処理してドキュメントを生成します
func ProcessFragments(db *sql.DB, dryRun bool, docIDOrTitle string) error {
	ctx := context.Background()

	// Gemini APIクライアントを初期化
	client, err := genai.NewClient(ctx, option.WithAPIKey(getGeminiAPIKey()))
	if err != nil {
		return fmt.Errorf("failed to create Gemini client: %w", err)
	}
	defer client.Close()

	// Gemini Pro モデルを取得
	model := client.GenerativeModel("gemini-1.5-flash")

	// 特定のドキュメント更新が指定されている場合
	if docIDOrTitle != "" {
		return updateSpecificDocument(ctx, db, model, docIDOrTitle, dryRun)
	}

	// 未処理のフラグメントを取得
	fragments, err := database.GetAllFragments(db, true) // unprocessed = true
	if err != nil {
		return fmt.Errorf("failed to get unprocessed fragments: %w", err)
	}

	if len(fragments) == 0 {
		fmt.Println("No unprocessed fragments found.")
		return nil
	}

	fmt.Printf("Found %d unprocessed fragments to process.\n", len(fragments))

	// フラグメントをグループ化（関連するものをまとめる）
	groups := groupRelatedFragments(fragments)
	fmt.Printf("Grouped fragments into %d potential documents.\n", len(groups))

	// 各グループを処理
	for i, group := range groups {
		fmt.Printf("\nProcessing group %d with %d fragments...\n", i+1, len(group))

		err := processFragmentGroupWithQuestions(ctx, db, model, group, dryRun)
		if err != nil {
			log.Printf("Error processing group %d: %v", i+1, err)
			continue
		}
	}

	return nil
}

// groupRelatedFragments は関連するフラグメントを意味的にグループ化します
func groupRelatedFragments(fragments []database.Fragment) [][]database.Fragment {
	if len(fragments) == 0 {
		return [][]database.Fragment{}
	}

	// AIを使って意味的グループ化を行う
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(getGeminiAPIKey()))
	if err != nil {
		log.Printf("Warning: failed to create Gemini client for grouping: %v", err)
		return fallbackTimeBasedGrouping(fragments)
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-1.5-flash")

	semanticGroups, err := performSemanticGrouping(ctx, model, fragments)
	if err != nil {
		log.Printf("Warning: failed to perform semantic grouping: %v", err)
		return fallbackTimeBasedGrouping(fragments)
	}

	return semanticGroups
}

// generateDocumentFromFragments はフラグメントからドキュメントを生成します
func generateDocumentFromFragments(ctx context.Context, model *genai.GenerativeModel, fragments []database.Fragment) (*GeneratedDocument, error) {
	// フラグメントを時系列順にソート（新しい順）
	sortedFragments := make([]database.Fragment, len(fragments))
	copy(sortedFragments, fragments)
	sort.Slice(sortedFragments, func(i, j int) bool {
		return sortedFragments[i].CreatedAt.After(sortedFragments[j].CreatedAt)
	})

	// フラグメントの内容を時系列情報付きで結合
	var fragmentContents []string
	for _, fragment := range sortedFragments {
		timestamp := fragment.CreatedAt.Format("2006-01-02 15:04:05")
		fragmentContents = append(fragmentContents, fmt.Sprintf("[%s] %s", timestamp, fragment.Content))
	}
	combinedContent := strings.Join(fragmentContents, "\n\n")

	// プロンプトを作成
	prompt := fmt.Sprintf(`以下の情報フラグメントを分析し、統合されたドキュメントを作成してください。

フラグメント（新しい順）:
%s

以下の形式で回答してください：
TITLE: [適切なタイトル]
SUMMARY: [2-3文の要約]
CONTENT: [マークダウン形式の詳細な内容]

指示：
- フラグメントは時系列順（新しい順）に並んでいます
- 新しい情報を優先し、古い情報と矛盾する場合は新しい情報を採用してください
- 情報の更新や変更がある場合は、最新の状態を反映してください
- 時系列の変化や進展があれば、それを構造的に表現してください
- マークダウン形式で読みやすく構造化してください
- タイトルは簡潔で内容を表すものにしてください`, combinedContent)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	if len(resp.Candidates) == 0 {
		return nil, fmt.Errorf("no response candidates received")
	}

	responseText := ""
	for _, part := range resp.Candidates[0].Content.Parts {
		responseText += fmt.Sprintf("%v", part)
	}

	// レスポンスをパースしてドキュメント構造に変換
	doc, err := parseGeneratedResponse(responseText)
	if err != nil {
		return nil, fmt.Errorf("failed to parse generated response: %w", err)
	}

	return doc, nil
}

// ExtractTags はコンテンツからAIを使ってタグを抽出します
func ExtractTags(content string) []string {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(getGeminiAPIKey()))
	if err != nil {
		log.Printf("Warning: failed to create Gemini client for tagging: %v", err)
		return fallbackKeywordExtraction(content)
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-1.5-flash")

	aiTags, err := extractTagsWithAI(ctx, model, content)
	if err != nil {
		log.Printf("Warning: failed to extract tags with AI: %v", err)
		return fallbackKeywordExtraction(content)
	}

	return aiTags
}

// fallbackKeywordExtraction はフォールバック用のキーワード抽出
func fallbackKeywordExtraction(content string) []string {
	keywords := []string{"Go", "JavaScript", "Python", "AI", "Machine Learning", "Web", "API", "Database", "Testing"}
	var tags []string

	contentLower := strings.ToLower(content)
	for _, keyword := range keywords {
		if strings.Contains(contentLower, strings.ToLower(keyword)) {
			tags = append(tags, keyword)
		}
	}
	return tags
}

// extractTagsWithAI はAIを使ってタグを抽出します
func extractTagsWithAI(ctx context.Context, model *genai.GenerativeModel, content string) ([]string, error) {
	prompt := fmt.Sprintf(`以下の文書内容から適切なタグを抽出してください。

文書内容:
%s

指示:
- 主要なトピック、技術、組織名、人物名、概念などを特定
- 各タグは1-3単語程度で簡潔に
- 最大5個までのタグを抽出
- 出力形式: tag1, tag2, tag3 のようにカンマ区切り

例:
paylab, AI, 業務効率化, NTTデータ, チーム管理`, content)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, err
	}

	if len(resp.Candidates) == 0 {
		return nil, fmt.Errorf("no response candidates received")
	}

	responseText := ""
	for _, part := range resp.Candidates[0].Content.Parts {
		responseText += fmt.Sprintf("%v", part)
	}

	return parseTagResponse(responseText), nil
}

// parseTagResponse はタグ抽出結果をパース
func parseTagResponse(response string) []string {
	// レスポンスをクリーンアップ
	response = strings.TrimSpace(response)

	var tags []string
	for _, tag := range strings.Split(response, ",") {
		tag = strings.TrimSpace(tag)
		if tag != "" && len(tag) <= 20 { // 長すぎるタグは除外
			tags = append(tags, tag)
		}
	}
	return tags
}

// parseGeneratedResponse はAIの応答をパースしてドキュメント構造に変換します
func parseGeneratedResponse(response string) (*GeneratedDocument, error) {
	lines := strings.Split(response, "\n")
	doc := &GeneratedDocument{}

	currentSection := ""
	var contentLines []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "TITLE:") {
			doc.Title = strings.TrimSpace(strings.TrimPrefix(trimmed, "TITLE:"))
			currentSection = "title"
		} else if strings.HasPrefix(trimmed, "SUMMARY:") {
			doc.Summary = strings.TrimSpace(strings.TrimPrefix(trimmed, "SUMMARY:"))
			currentSection = "summary"
		} else if strings.HasPrefix(trimmed, "CONTENT:") {
			currentSection = "content"
			content := strings.TrimSpace(strings.TrimPrefix(trimmed, "CONTENT:"))
			if content != "" {
				contentLines = append(contentLines, content)
			}
		} else if currentSection == "content" {
			contentLines = append(contentLines, line)
		} else if currentSection == "summary" && trimmed != "" {
			doc.Summary += " " + trimmed
		}
	}

	doc.Content = strings.Join(contentLines, "\n")

	// デフォルト値を設定
	if doc.Title == "" {
		doc.Title = "Generated Document"
	}
	if doc.Summary == "" {
		doc.Summary = "AI generated document from fragments"
	}

	return doc, nil
}

// updateSpecificDocument は特定のドキュメントを更新します
func updateSpecificDocument(ctx context.Context, db *sql.DB, model *genai.GenerativeModel, docIDOrTitle string, dryRun bool) error {
	// 既存のドキュメントを取得
	doc, err := database.GetDocumentByIDOrTitle(db, docIDOrTitle)
	if err != nil {
		return fmt.Errorf("failed to get document: %w", err)
	}

	// そのドキュメントに関連するフラグメントを取得
	fragments, err := database.GetDocumentFragments(db, doc.ID)
	if err != nil {
		return fmt.Errorf("failed to get document fragments: %w", err)
	}

	if len(fragments) == 0 {
		fmt.Printf("No fragments found for document '%s'.\n", docIDOrTitle)
		return nil
	}

	fmt.Printf("Updating document '%s' with %d fragments.\n", doc.Title, len(fragments))

	// ドキュメントを再生成
	newDoc, err := generateDocumentFromFragments(ctx, model, fragments)
	if err != nil {
		return fmt.Errorf("failed to regenerate document: %w", err)
	}

	if dryRun {
		fmt.Println("\n--- Document Update Preview ---")
		fmt.Printf("Title: %s\n", newDoc.Title)
		fmt.Printf("Summary: %s\n", newDoc.Summary)
		fmt.Printf("Content:\n%s\n", newDoc.Content)
		fmt.Println("--- End Preview ---")
		return nil
	}

	// ドキュメントを更新
	err = database.UpdateDocument(db, doc.ID, newDoc.Title, newDoc.Content, newDoc.Summary)
	if err != nil {
		return fmt.Errorf("failed to update document: %w", err)
	}

	fmt.Printf("Document '%s' updated successfully.\n", doc.Title)
	return nil
}

// processFragmentGroup はフラグメントのグループを処理してドキュメントを生成します
func processFragmentGroup(ctx context.Context, db *sql.DB, model *genai.GenerativeModel, fragments []database.Fragment, dryRun bool) error {
	if dryRun {
		fmt.Println("\n--- Processing Preview ---")
		fmt.Printf("Would create new document from %d fragments\n", len(fragments))
		fmt.Println("--- End Preview ---")
		return nil
	}

	// 新しいドキュメントを作成
	generatedDoc, err := generateDocumentFromFragments(ctx, model, fragments)
	if err != nil {
		return fmt.Errorf("failed to generate document: %w", err)
	}

	// ドキュメントをデータベースに保存
	docID, err := database.InsertDocument(db, generatedDoc.Title, generatedDoc.Content, generatedDoc.Summary)
	if err != nil {
		return fmt.Errorf("failed to insert document: %w", err)
	}

	// フラグメントをリンク
	for _, fragment := range fragments {
		err = database.LinkFragmentToDocument(db, fragment.ID, docID)
		if err != nil {
			log.Printf("Warning: failed to link fragment %d to document %d: %v", fragment.ID, docID, err)
		}
	}

	// タグを抽出して追加
	tags := ExtractTags(generatedDoc.Content)
	for _, tag := range tags {
		tagID, err := database.InsertOrGetTag(db, tag)
		if err != nil {
			log.Printf("Warning: failed to insert tag '%s': %v", tag, err)
			continue
		}
		err = database.LinkDocumentToTag(db, docID, tagID)
		if err != nil {
			log.Printf("Warning: failed to link document to tag '%s': %v", tag, err)
		}
	}

	fmt.Printf("New document '%s' created successfully with ID %d.\n", generatedDoc.Title, docID)
	return nil
}

// fallbackTimeBasedGrouping は時系列ベースのフォールバックグループ化
func fallbackTimeBasedGrouping(fragments []database.Fragment) [][]database.Fragment {
	if len(fragments) == 0 {
		return [][]database.Fragment{}
	}

	// 簡単な実装：すべてのフラグメントを一つのグループに
	return [][]database.Fragment{fragments}
}

// performSemanticGrouping はAIを使って意味的グループ化を実行
func performSemanticGrouping(ctx context.Context, model *genai.GenerativeModel, fragments []database.Fragment) ([][]database.Fragment, error) {
	if len(fragments) <= 1 {
		return [][]database.Fragment{fragments}, nil
	}

	// 簡単な実装：関連性の高いフラグメントを検出
	// より複雑な実装は後で追加可能
	return [][]database.Fragment{fragments}, nil
}

// processFragmentGroupWithQuestions は質問生成を含むフラグメントグループ処理
func processFragmentGroupWithQuestions(ctx context.Context, db *sql.DB, model *genai.GenerativeModel, fragments []database.Fragment, dryRun bool) error {
	if dryRun {
		fmt.Println("\n--- Processing Preview ---")
		fmt.Printf("Would create new document from %d fragments\n", len(fragments))
		fmt.Println("--- End Preview ---")
		return nil
	}

	// フラグメントを分析して質問を生成
	questions, err := generateQuestionsForFragments(ctx, model, fragments)
	if err != nil {
		log.Printf("Warning: failed to generate questions: %v", err)
	}

	// 質問をデータベースに保存
	var fragmentIDs []int
	for _, fragment := range fragments {
		fragmentIDs = append(fragmentIDs, fragment.ID)
	}

	for _, question := range questions {
		questionID, err := database.InsertQuestion(db, question, fragmentIDs, nil)
		if err != nil {
			log.Printf("Warning: failed to insert question '%s': %v", question, err)
		} else {
			fmt.Printf("Generated question (ID %d): %s\n", questionID, question)
		}
	}

	// 通常のドキュメント生成も実行
	return processFragmentGroup(ctx, db, model, fragments, dryRun)
}

// generateQuestionsForFragments はフラグメントに対する質問を生成します
func generateQuestionsForFragments(ctx context.Context, model *genai.GenerativeModel, fragments []database.Fragment) ([]string, error) {
	if len(fragments) == 0 {
		return nil, nil
	}

	// フラグメントの内容を結合
	var fragmentContents []string
	for _, fragment := range fragments {
		fragmentContents = append(fragmentContents, fragment.Content)
	}
	combinedContent := strings.Join(fragmentContents, "\n")

	// 質問生成プロンプト
	prompt := fmt.Sprintf(`以下の情報フラグメントを分析し、ドキュメント作成時に明確にすべき曖昧性や矛盾を特定して質問を生成してください。

情報フラグメント:
%s

指示:
- 人物や組織の同定に関する曖昧性（例：「斎藤さん」と「saikyoさん」は同一人物？）
- 時期や日付の不明確さ（例：「最近」「先日」の具体的な日時）
- 場所や部署名の曖昧性（例：「あの部署」「向こうのチーム」の具体名）
- 専門用語や略語の定義（例：「例の件」「あのシステム」の具体的内容）
- 数値や仕様の不明確さ（例：「多くの」「大幅な」の具体的な値）
- 最大3個の質問を生成
- 各質問は1行で、質問文のみを出力
- 出力形式: 各質問を改行で区切る

例:
「斎藤さん」と「saikyoさん」は同一人物ですか？
「最近開始した」プロジェクトの正確な開始日はいつですか？
「あのシステム」とは具体的にどのシステムを指していますか？`, combinedContent)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, fmt.Errorf("failed to generate questions: %w", err)
	}

	if len(resp.Candidates) == 0 {
		return nil, fmt.Errorf("no response candidates received")
	}

	responseText := ""
	for _, part := range resp.Candidates[0].Content.Parts {
		responseText += fmt.Sprintf("%v", part)
	}

	// レスポンスを質問リストに分割
	questions := strings.Split(strings.TrimSpace(responseText), "\n")
	var cleanQuestions []string
	for _, q := range questions {
		q = strings.TrimSpace(q)
		if q != "" && strings.HasSuffix(q, "？") || strings.HasSuffix(q, "?") {
			cleanQuestions = append(cleanQuestions, q)
		}
	}

	return cleanQuestions, nil
}

// AnswerQuestionWithDocuments はドキュメントを元に質問に回答します
func AnswerQuestionWithDocuments(db *sql.DB, question string) (string, error) {
	ctx := context.Background()

	// 全ドキュメントを取得
	documents, err := database.GetAllDocuments(db)
	if err != nil {
		return "", fmt.Errorf("failed to get documents: %w", err)
	}

	if len(documents) == 0 {
		return "回答できません。参照可能なドキュメントがありません。", nil
	}

	// Gemini APIクライアントを初期化
	client, err := genai.NewClient(ctx, option.WithAPIKey(getGeminiAPIKey()))
	if err != nil {
		return "", fmt.Errorf("failed to create Gemini client: %w", err)
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-1.5-flash")

	// ドキュメント内容を結合
	var documentContents []string
	for _, doc := range documents {
		summary := "要約なし"
		if doc.Summary.Valid {
			summary = doc.Summary.String
		}
		documentContents = append(documentContents, fmt.Sprintf("【%s】\n要約: %s\n内容: %s", doc.Title, summary, doc.Content))
	}
	combinedContent := strings.Join(documentContents, "\n\n---\n\n")

	// 回答生成プロンプト
	prompt := fmt.Sprintf(`以下のドキュメントを参照して、質問に回答してください。

質問: %s

参照ドキュメント:
%s

指示:
- ドキュメントの内容のみを根拠として回答してください
- 推測や一般的な知識ではなく、提供されたドキュメントに基づいた回答を心がけてください
- ドキュメントに該当する情報がない場合は「提供されたドキュメントには該当する情報がありません」と回答してください
- 回答の根拠となったドキュメント名を明記してください
- 簡潔で分かりやすい回答を提供してください

回答形式:
【回答】
[具体的な回答内容]

【参考ドキュメント】
- [ドキュメント名1]
- [ドキュメント名2]`, question, combinedContent)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("failed to generate answer: %w", err)
	}

	if len(resp.Candidates) == 0 {
		return "", fmt.Errorf("no response candidates received")
	}

	responseText := ""
	for _, part := range resp.Candidates[0].Content.Parts {
		responseText += fmt.Sprintf("%v", part)
	}

	return strings.TrimSpace(responseText), nil
}

// SearchAndAnswerQuestion は関連ドキュメントを検索して質問に回答します
func SearchAndAnswerQuestion(db *sql.DB, question string, searchTerm string) (string, error) {
	ctx := context.Background()

	// 検索語が指定されている場合は検索、そうでなければ全ドキュメント
	var documents []database.Document
	var err error

	if searchTerm != "" {
		documents, err = database.SearchDocuments(db, searchTerm)
		if err != nil {
			return "", fmt.Errorf("failed to search documents: %w", err)
		}
	} else {
		documents, err = database.GetAllDocuments(db)
		if err != nil {
			return "", fmt.Errorf("failed to get documents: %w", err)
		}
	}

	if len(documents) == 0 {
		if searchTerm != "" {
			return fmt.Sprintf("検索語「%s」に関連するドキュメントが見つかりません。", searchTerm), nil
		} else {
			return "回答できません。参照可能なドキュメントがありません。", nil
		}
	}

	// Gemini APIクライアントを初期化
	client, err := genai.NewClient(ctx, option.WithAPIKey(getGeminiAPIKey()))
	if err != nil {
		return "", fmt.Errorf("failed to create Gemini client: %w", err)
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-1.5-flash")

	// ドキュメント内容を結合
	var documentContents []string
	for _, doc := range documents {
		summary := "要約なし"
		if doc.Summary.Valid {
			summary = doc.Summary.String
		}
		documentContents = append(documentContents, fmt.Sprintf("【%s】\n要約: %s\n内容: %s", doc.Title, summary, doc.Content))
	}
	combinedContent := strings.Join(documentContents, "\n\n---\n\n")

	// 回答生成プロンプト
	searchInfo := ""
	if searchTerm != "" {
		searchInfo = fmt.Sprintf("（検索語「%s」で%d件のドキュメントを検索）", searchTerm, len(documents))
	}

	prompt := fmt.Sprintf(`以下のドキュメントを参照して、質問に回答してください。%s

質問: %s

参照ドキュメント:
%s

指示:
- ドキュメントの内容のみを根拠として回答してください
- 推測や一般的な知識ではなく、提供されたドキュメントに基づいた回答を心がけてください
- ドキュメントに該当する情報がない場合は「提供されたドキュメントには該当する情報がありません」と回答してください
- 回答の根拠となったドキュメント名を明記してください
- 複数のドキュメントから情報を統合して回答しても構いません
- 簡潔で分かりやすい回答を提供してください

回答形式:
【回答】
[具体的な回答内容]

【参考ドキュメント】
- [ドキュメント名1]
- [ドキュメント名2]`, searchInfo, question, combinedContent)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("failed to generate answer: %w", err)
	}

	if len(resp.Candidates) == 0 {
		return "", fmt.Errorf("no response candidates received")
	}

	responseText := ""
	for _, part := range resp.Candidates[0].Content.Parts {
		responseText += fmt.Sprintf("%v", part)
	}

	return strings.TrimSpace(responseText), nil
}
