package cli

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"insight/internal/ai"
	"insight/internal/database"
	"insight/internal/export"
	"insight/internal/web"
	"log"
	"os"
)

type App struct {
	db *sql.DB
}

func NewApp(db *sql.DB) *App {
	return &App{db: db}
}

func (app *App) Run(args []string) error {
	if len(args) < 2 {
		fmt.Println("Usage: insight <command> [arguments]")
		return fmt.Errorf("no command provided")
	}

	switch args[1] {
	case "add":
		return app.handleAddCommand(args[2:])
	case "process":
		return app.handleProcessCommand(args[2:])
	case "doc":
		return app.handleDocCommand(args[2:])
	case "tag":
		return app.handleTagCommand(args[2:])
	case "find":
		return app.handleFindCommand(args[2:])
	case "status":
		return app.handleStatusCommand(args[2:])
	case "fragment":
		return app.handleFragmentCommand(args[2:])
	case "web":
		return app.handleWebCommand(args[2:])
	case "reset":
		return app.handleResetCommand(args[2:])
	case "clear":
		return app.handleClearCommand(args[2:])
	case "export":
		return app.handleExportCommand(args[2:])
	case "question":
		return app.handleQuestionCommand(args[2:])
	case "answer":
		return app.handleAnswerCommand(args[2:])
	case "ask":
		return app.handleAskCommand(args[2:])
	default:
		return fmt.Errorf("unknown command: %s", args[1])
	}
}

func (app *App) handleAddCommand(args []string) error {
	addCmd := flag.NewFlagSet("add", flag.ExitOnError)
	var url string
	var imagePath string
	addCmd.StringVar(&url, "url", "", "URL for the fragment")
	addCmd.StringVar(&imagePath, "image", "", "Path to an image for the fragment")

	addCmd.Parse(args)

	nonFlagArgs := addCmd.Args()
	var text string
	if len(nonFlagArgs) > 0 {
		text = nonFlagArgs[0]
	}

	fragmentContent := make(map[string]string)
	if text != "" {
		fragmentContent["text"] = text
	}
	if url != "" {
		fragmentContent["url"] = url
	}
	if imagePath != "" {
		fragmentContent["image_path"] = imagePath
	}

	if len(fragmentContent) == 0 {
		fmt.Println("Usage: insight add [text] [--url <url>] [--image <path>]")
		os.Exit(1)
	}

	jsonContent, err := json.Marshal(fragmentContent)
	if err != nil {
		return fmt.Errorf("failed to marshal fragment content to JSON: %w", err)
	}

	id, err := database.InsertFragment(app.db, string(jsonContent))
	if err != nil {
		return fmt.Errorf("failed to insert fragment: %w", err)
	}

	fmt.Printf("Fragment added with ID: %d\n", id)
	return nil
}

func (app *App) handleProcessCommand(args []string) error {
	processCmd := flag.NewFlagSet("process", flag.ExitOnError)
	var dryRun bool
	var docIDOrTitle string
	var enableAI bool
	processCmd.BoolVar(&dryRun, "dry-run", false, "Preview changes without modifying the database")
	processCmd.StringVar(&docIDOrTitle, "doc", "", "Force update on a specific document")
	processCmd.BoolVar(&enableAI, "ai", false, "Enable AI processing of fragments")

	processCmd.Parse(args)

	fmt.Println("Process command executed.")
	if dryRun {
		fmt.Println("Dry run mode enabled.")
	}
	if docIDOrTitle != "" {
		fmt.Printf("Processing document: %s\n", docIDOrTitle)
	}

	// AI処理が有効な場合のみ実行
	if enableAI {
		err := ai.ProcessFragments(app.db, dryRun, docIDOrTitle)
		if err != nil {
			return fmt.Errorf("AI processing failed: %w", err)
		}
	}
	return nil
}

func (app *App) handleAskCommand(args []string) error {
	askCmd := flag.NewFlagSet("ask", flag.ExitOnError)
	var search string
	askCmd.StringVar(&search, "search", "", "Search term to filter documents before answering")
	askCmd.Parse(args)

	nonFlagArgs := askCmd.Args()
	if len(nonFlagArgs) == 0 {
		fmt.Println("Usage: insight ask \"質問内容\" [--search \"検索語\"]")
		os.Exit(1)
	}

	question := nonFlagArgs[0]

	fmt.Printf("質問: %s\n", question)
	if search != "" {
		fmt.Printf("検索範囲: \"%s\"\n", search)
	}
	fmt.Println("\n回答を生成中...")

	var answer string
	var err error

	if search != "" {
		answer, err = ai.SearchAndAnswerQuestion(app.db, question, search)
	} else {
		answer, err = ai.AnswerQuestionWithDocuments(app.db, question)
	}

	if err != nil {
		return fmt.Errorf("failed to generate answer: %w", err)
	}

	fmt.Printf("\n%s\n", answer)
	return nil
}

func (app *App) handleClearCommand(args []string) error {
	clearCmd := flag.NewFlagSet("clear", flag.ExitOnError)
	var confirm bool
	clearCmd.BoolVar(&confirm, "confirm", false, "Confirm deletion of all data")
	clearCmd.Parse(args)

	if !confirm {
		fmt.Println("このコマンドは全てのデータ（フラグメント、ドキュメント、質問）を削除します。")
		fmt.Println("実行するには --confirm フラグを付けてください:")
		fmt.Println("  insight clear --confirm")
		return nil
	}

	fmt.Println("全てのデータを削除しています...")
	err := database.DeleteAllData(app.db)
	if err != nil {
		return fmt.Errorf("failed to delete all data: %w", err)
	}

	fmt.Println("データベースが完全にクリアされました。")
	fmt.Println("新しいフラグメントを追加して本格的に利用を開始してください！")
	return nil
}

func (app *App) handleQuestionCommand(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("question subcommand required")
	}

	switch args[0] {
	case "list":
		return app.handleQuestionListCommand(args[1:])
	default:
		return fmt.Errorf("unknown question subcommand: %s", args[0])
	}
}

func (app *App) handleQuestionListCommand(args []string) error {
	questionListCmd := flag.NewFlagSet("question list", flag.ExitOnError)
	var pending bool
	questionListCmd.BoolVar(&pending, "pending", false, "Show only pending questions")
	questionListCmd.Parse(args)

	var questions []database.Question
	var err error

	if pending {
		questions, err = database.GetPendingQuestions(app.db)
		if err != nil {
			return fmt.Errorf("failed to get pending questions: %w", err)
		}
	} else {
		questions, err = database.GetAllQuestions(app.db)
		if err != nil {
			return fmt.Errorf("failed to get questions: %w", err)
		}
	}

	if len(questions) == 0 {
		if pending {
			fmt.Println("No pending questions found.")
		} else {
			fmt.Println("No questions found.")
		}
		return nil
	}

	fmt.Println("ID\tStatus\tQuestion\tCreated At")
	fmt.Println("--\t------\t--------\t----------")
	for _, q := range questions {
		fmt.Printf("%d\t%s\t%s\t%s\n", q.ID, q.Status, q.QuestionText, q.CreatedAt.Format("2006-01-02 15:04:05"))
	}
	return nil
}

func (app *App) handleAnswerCommand(args []string) error {
	if len(args) < 2 {
		fmt.Println("Usage: insight answer <question_id> <answer_text>")
		os.Exit(1)
	}

	questionID := args[0]
	answerText := args[1]

	// 質問IDを数値に変換
	var qID int
	if _, err := fmt.Sscanf(questionID, "%d", &qID); err != nil {
		return fmt.Errorf("invalid question ID: %s", questionID)
	}

	// 質問の内容を取得
	question, err := database.GetQuestionByID(app.db, qID)
	if err != nil {
		return fmt.Errorf("failed to get question: %w", err)
	}

	// 回答をフラグメントとして保存（質問内容も含める）
	fragmentContent := map[string]string{
		"text":        answerText,
		"type":        "answer",
		"question":    question.QuestionText,
		"question_id": fmt.Sprintf("%d", qID),
	}

	jsonContent, err := json.Marshal(fragmentContent)
	if err != nil {
		return fmt.Errorf("failed to marshal answer content to JSON: %w", err)
	}

	fragmentID, err := database.InsertFragment(app.db, string(jsonContent))
	if err != nil {
		return fmt.Errorf("failed to insert answer fragment: %w", err)
	}

	// 質問に回答を関連付け
	err = database.AnswerQuestion(app.db, qID, fragmentID)
	if err != nil {
		return fmt.Errorf("failed to answer question: %w", err)
	}

	fmt.Printf("Question %d answered successfully with fragment ID %d\n", qID, fragmentID)
	return nil
}

func (app *App) handleDocCommand(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("doc subcommand required")
	}

	switch args[0] {
	case "list":
		return app.handleDocListCommand(args[1:])
	case "show":
		return app.handleDocShowCommand(args[1:])
	default:
		return fmt.Errorf("unknown doc subcommand: %s", args[0])
	}
}

func (app *App) handleDocListCommand(args []string) error {
	documents, err := database.GetAllDocuments(app.db)
	if err != nil {
		return fmt.Errorf("failed to get documents: %w", err)
	}

	if len(documents) == 0 {
		fmt.Println("No documents found.")
		return nil
	}

	fmt.Println("ID\tTitle\tLast Updated")
	fmt.Println("--\t-----\t------------")
	for _, doc := range documents {
		fmt.Printf("%d\t%s\t%s\n", doc.ID, doc.Title, doc.UpdatedAt.Format("2006-01-02 15:04:05"))
	}
	return nil
}

func (app *App) handleDocShowCommand(args []string) error {
	if len(args) == 0 {
		fmt.Println("Usage: insight doc show <id_or_title>")
		os.Exit(1)
	}

	idOrTitle := args[0]
	doc, err := database.GetDocumentByIDOrTitle(app.db, idOrTitle)
	if err != nil {
		log.Fatalf("Failed to get document: %v", err)
	}

	// ドキュメントの基本情報を表示
	fmt.Printf("# %s\n\n", doc.Title)
	fmt.Printf("**ID:** %d\n", doc.ID)
	fmt.Printf("**Created:** %s\n", doc.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("**Updated:** %s\n\n", doc.UpdatedAt.Format("2006-01-02 15:04:05"))

	// 要約があれば表示
	if doc.Summary.Valid && doc.Summary.String != "" {
		fmt.Printf("**Summary:** %s\n\n", doc.Summary.String)
	}

	// タグを表示
	tags, err := database.GetDocumentTags(app.db, doc.ID)
	if err != nil {
		log.Printf("Warning: Failed to get document tags: %v", err)
	} else if len(tags) > 0 {
		fmt.Printf("**Tags:** ")
		for i, tag := range tags {
			if i > 0 {
				fmt.Printf(", ")
			}
			fmt.Printf("%s", tag)
		}
		fmt.Printf("\n\n")
	}

	// メインコンテンツを表示
	fmt.Printf("## Content\n\n%s\n\n", doc.Content)

	// 関連フラグメントを表示
	fragments, err := database.GetDocumentFragments(app.db, doc.ID)
	if err != nil {
		log.Printf("Warning: Failed to get document fragments: %v", err)
	} else if len(fragments) > 0 {
		fmt.Printf("## Source Fragments\n\n")
		for _, fragment := range fragments {
			fmt.Printf("- **Fragment %d** (%s): %s\n", fragment.ID, fragment.CreatedAt.Format("2006-01-02 15:04:05"), fragment.Content)
		}
	}
	return nil
}

func (app *App) handleTagCommand(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("tag subcommand required")
	}

	switch args[0] {
	case "list":
		return app.handleTagListCommand(args[1:])
	default:
		return fmt.Errorf("unknown tag subcommand: %s", args[0])
	}
}

func (app *App) handleTagListCommand(args []string) error {
	tags, err := database.GetAllTags(app.db)
	if err != nil {
		return fmt.Errorf("failed to get tags: %w", err)
	}

	if len(tags) == 0 {
		fmt.Println("No tags found.")
		return nil
	}

	fmt.Println("ID\tTag\tCount")
	fmt.Println("--\t-----\t-----")
	for _, tag := range tags {
		fmt.Printf("%d\t%s\t%d\n", tag.ID, tag.Name, tag.Count)
	}
	return nil
}

func (app *App) handleFindCommand(args []string) error {
	findCmd := flag.NewFlagSet("find", flag.ExitOnError)
	var tag string
	var query string
	findCmd.StringVar(&tag, "tag", "", "Tag to search for")
	findCmd.StringVar(&query, "query", "", "Search query for full-text search")
	findCmd.Parse(args)

	if tag != "" {
		return app.handleFindTagCommand(tag)
	} else if query != "" {
		return app.handleFindQueryCommand(query)
	} else {
		fmt.Println("Usage: insight find --tag <tag_name> | --query \"<search_term>\"")
		os.Exit(1)
	}
	return nil
}

func (app *App) handleFindTagCommand(tagName string) error {
	documents, err := database.GetDocumentsByTag(app.db, tagName)
	if err != nil {
		return fmt.Errorf("failed to find documents by tag: %w", err)
	}

	if len(documents) == 0 {
		fmt.Printf("No documents found with tag: %s\n", tagName)
		return nil
	}

	fmt.Printf("Found %d document(s) with tag '%s':\n\n", len(documents), tagName)
	fmt.Println("ID\tTitle\tLast Updated")
	fmt.Println("--\t-----\t------------")
	for _, doc := range documents {
		fmt.Printf("%d\t%s\t%s\n", doc.ID, doc.Title, doc.UpdatedAt.Format("2006-01-02 15:04:05"))
	}
	return nil
}

func (app *App) handleFindQueryCommand(searchTerm string) error {
	documents, err := database.SearchDocuments(app.db, searchTerm)
	if err != nil {
		return fmt.Errorf("failed to search documents: %w", err)
	}

	if len(documents) == 0 {
		fmt.Printf("No documents found matching query: %s\n", searchTerm)
		return nil
	}

	fmt.Printf("Found %d document(s) matching '%s':\n\n", len(documents), searchTerm)
	fmt.Println("ID\tTitle\tLast Updated")
	fmt.Println("--\t-----\t------------")
	for _, doc := range documents {
		fmt.Printf("%d\t%s\t%s\n", doc.ID, doc.Title, doc.UpdatedAt.Format("2006-01-02 15:04:05"))
	}
	return nil
}

func (app *App) handleStatusCommand(args []string) error {
	docCount, totalFragmentCount, unprocessedFragmentCount, err := database.GetCounts(app.db)
	if err != nil {
		return fmt.Errorf("failed to get counts: %w", err)
	}

	fmt.Printf("Documents: %d\n", docCount)
	fmt.Printf("Total Fragments: %d\n", totalFragmentCount)
	fmt.Printf("Unprocessed Fragments: %d\n", unprocessedFragmentCount)
	return nil
}

func (app *App) handleFragmentCommand(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("fragment subcommand required")
	}

	switch args[0] {
	case "list":
		return app.handleFragmentListCommand(args[1:])
	default:
		return fmt.Errorf("unknown fragment subcommand: %s", args[0])
	}
}

func (app *App) handleFragmentListCommand(args []string) error {
	fragmentListCmd := flag.NewFlagSet("fragment list", flag.ExitOnError)
	var unprocessed bool
	fragmentListCmd.BoolVar(&unprocessed, "unprocessed", false, "Show only fragments that have not yet been processed into a document.")
	fragmentListCmd.Parse(args)

	fragments, err := database.GetAllFragments(app.db, unprocessed)
	if err != nil {
		return fmt.Errorf("failed to get fragments: %w", err)
	}

	if len(fragments) == 0 {
		fmt.Println("No fragments found.")
		return nil
	}

	fmt.Println("ID\tContent\tCreated At")
	fmt.Println("--\t-------\t----------")
	for _, fragment := range fragments {
		fmt.Printf("%d\t%s\t%s\n", fragment.ID, fragment.Content, fragment.CreatedAt.Format("2006-01-02 15:04:05"))
	}
	return nil
}

func (app *App) handleWebCommand(args []string) error {
	webCmd := flag.NewFlagSet("web", flag.ExitOnError)
	var port string
	webCmd.StringVar(&port, "port", "8080", "Port to run the web server on")
	webCmd.Parse(args)

	server := web.NewServer(app.db)
	server.Start(port)
	return nil
}

func (app *App) handleResetCommand(args []string) error {
	resetCmd := flag.NewFlagSet("reset", flag.ExitOnError)
	var confirm bool
	resetCmd.BoolVar(&confirm, "confirm", false, "Confirm deletion of all documents")
	resetCmd.Parse(args)

	if !confirm {
		fmt.Println("このコマンドは全てのドキュメントを削除し、フラグメントから再構築します。")
		fmt.Println("実行するには --confirm フラグを付けてください:")
		fmt.Println("  insight reset --confirm")
		return nil
	}

	fmt.Println("全てのドキュメントを削除しています...")
	err := database.DeleteAllDocuments(app.db)
	if err != nil {
		return fmt.Errorf("failed to delete documents: %w", err)
	}

	fmt.Println("削除完了。フラグメントから再構築を開始します...")
	// TODO: AI処理を実装
	fmt.Println("AI processing not yet implemented in refactored version")
	fmt.Println("再構築完了！")
	return nil
}

func (app *App) handleExportCommand(args []string) error {
	exportCmd := flag.NewFlagSet("export", flag.ExitOnError)
	var outputDir string
	exportCmd.StringVar(&outputDir, "dir", "./exports", "Export directory")
	exportCmd.Parse(args)

	err := export.ExportAllDocumentsToSeparateFiles(app.db, outputDir)
	if err != nil {
		return fmt.Errorf("failed to export documents: %w", err)
	}
	return nil
}
