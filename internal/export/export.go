package export

import (
	"database/sql"
	"fmt"
	"insight/internal/database"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ExportAllDocumentsToSeparateFiles は各ドキュメントを個別のMarkdownファイルとしてエクスポートします
func ExportAllDocumentsToSeparateFiles(db *sql.DB, outputDir string) error {
	// 出力ディレクトリの作成
	err := os.MkdirAll(outputDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// 全ドキュメントを取得
	documents, err := database.GetAllDocuments(db)
	if err != nil {
		return fmt.Errorf("failed to get documents: %w", err)
	}

	if len(documents) == 0 {
		fmt.Println("No documents to export.")
		return nil
	}

	fmt.Printf("Exporting %d documents to separate files in %s...\n", len(documents), outputDir)

	// 各ドキュメントを個別ファイルとしてエクスポート
	for _, doc := range documents {
		err := exportDocumentToFile(db, doc, outputDir)
		if err != nil {
			fmt.Printf("Warning: failed to export document %d: %v\n", doc.ID, err)
			continue
		}
		fmt.Printf("- Exported: %s\n", sanitizeFilename(doc.Title)+".md")
	}

	fmt.Printf("Export completed! %d files created in %s\n", len(documents), outputDir)
	return nil
}

// exportDocumentToFile は単一ドキュメントをファイルにエクスポートします
func exportDocumentToFile(db *sql.DB, doc database.Document, outputDir string) error {
	// ファイル名を生成（安全なファイル名に変換）
	filename := sanitizeFilename(doc.Title) + ".md"
	filepath := filepath.Join(outputDir, filename)

	// ファイルを作成
	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Markdownコンテンツを生成
	content, err := generateMarkdownContent(db, doc)
	if err != nil {
		return fmt.Errorf("failed to generate markdown content: %w", err)
	}

	// ファイルに書き込み
	_, err = file.WriteString(content)
	if err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	return nil
}

// generateMarkdownContent はドキュメントのMarkdownコンテンツを生成します
func generateMarkdownContent(db *sql.DB, doc database.Document) (string, error) {
	var content strings.Builder

	// タイトル
	content.WriteString(fmt.Sprintf("# %s\n\n", doc.Title))

	// メタデータ
	content.WriteString("## Document Information\n\n")
	content.WriteString(fmt.Sprintf("- **ID:** %d\n", doc.ID))
	content.WriteString(fmt.Sprintf("- **Created:** %s\n", doc.CreatedAt.Format("2006-01-02 15:04:05")))
	content.WriteString(fmt.Sprintf("- **Updated:** %s\n", doc.UpdatedAt.Format("2006-01-02 15:04:05")))

	// 要約
	if doc.Summary.Valid && doc.Summary.String != "" {
		content.WriteString(fmt.Sprintf("- **Summary:** %s\n", doc.Summary.String))
	}

	// タグ
	tags, err := database.GetDocumentTags(db, doc.ID)
	if err == nil && len(tags) > 0 {
		content.WriteString(fmt.Sprintf("- **Tags:** %s\n", strings.Join(tags, ", ")))
	}

	content.WriteString("\n")

	// メインコンテンツ
	content.WriteString("## Content\n\n")
	content.WriteString(doc.Content)
	content.WriteString("\n\n")

	// ソースフラグメント
	fragments, err := database.GetDocumentFragments(db, doc.ID)
	if err == nil && len(fragments) > 0 {
		content.WriteString("## Source Fragments\n\n")
		for _, fragment := range fragments {
			content.WriteString(fmt.Sprintf("- **Fragment %d** (%s): %s\n",
				fragment.ID,
				fragment.CreatedAt.Format("2006-01-02 15:04:05"),
				fragment.Content))
		}
		content.WriteString("\n")
	}

	return content.String(), nil
}

// sanitizeFilename はファイル名に使用できない文字を取り除きます
func sanitizeFilename(filename string) string {
	// 危険な文字を削除または置換
	reg := regexp.MustCompile(`[<>:"/\\|?*]`)
	filename = reg.ReplaceAllString(filename, "_")

	// 連続するスペースやアンダースコアを単一に
	reg = regexp.MustCompile(`[\s_]+`)
	filename = reg.ReplaceAllString(filename, "_")

	// 先頭と末尾の不要な文字を削除
	filename = strings.Trim(filename, "._- ")

	// 長すぎる場合は切り詰め
	if len(filename) > 100 {
		filename = filename[:100]
	}

	// 空の場合はデフォルト名
	if filename == "" {
		filename = "untitled"
	}

	return filename
}
