package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"insight/src/ai"
	"insight/src/db"
	"insight/src/usecase"

	"github.com/urfave/cli/v3"
)

func main() {
	app := &cli.Command{
		Name:  "insight",
		Usage: "A tool for managing fragments and documents",
		Commands: []*cli.Command{
			{
				Name:  "fragment",
				Usage: "Fragment operations",
				Commands: []*cli.Command{
					{
						Name:  "create",
						Usage: "Create a new fragment",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "content",
								Aliases:  []string{"c"},
								Usage:    "Content of the fragment",
								Required: true,
							},
						},
						Action: createFragment,
					},
					{
						Name:  "list",
						Usage: "List all fragments",
						Action: listFragments,
					},
					{
						Name:  "delete",
						Usage: "Delete a fragment",
						Flags: []cli.Flag{
							&cli.IntFlag{
								Name:     "id",
								Usage:    "Fragment ID to delete",
								Required: true,
							},
						},
						Action: deleteFragment,
					},
				},
			},
			{
				Name:  "document",
				Usage: "Document operations",
				Commands: []*cli.Command{
					{
						Name:   "list",
						Usage:  "List all documents",
						Action: listDocuments,
					},
					{
						Name:  "show",
						Usage: "Show document details",
						Flags: []cli.Flag{
							&cli.IntFlag{
								Name:     "id",
								Usage:    "Document ID",
								Required: true,
							},
						},
						Action: showDocument,
					},
				},
			},
			{
				Name:  "ai",
				Usage: "AI operations",
				Commands: []*cli.Command{
					{
						Name:   "create",
						Usage:  "Create documents from fragments using AI",
						Action: createDocuments,
					},
					{
						Name:   "compress",
						Usage:  "Compress fragments by merging similar ones and removing low-value ones",
						Action: compressFragments,
					},
				},
			},
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

func createFragment(ctx context.Context, c *cli.Command) error {
	content := c.String("content")

	// データベース初期化
	database, err := db.Init(nil)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer db.Close(database)

	// ユースケース初期化
	fragmentUsecase := usecase.NewFragmentUsecase(database)

	// Fragment作成
	input := usecase.CreateFragmentInput{
		Content: content,
	}

	fragment, err := fragmentUsecase.CreateFragment(input)
	if err != nil {
		return fmt.Errorf("failed to create fragment: %w", err)
	}

	fmt.Printf("Fragment created successfully!\n")
	fmt.Printf("ID: %d\n", fragment.ID)
	fmt.Printf("Content: %s\n", fragment.Content)
	fmt.Printf("Created at: %s\n", fragment.CreatedAt.Format("2006-01-02 15:04:05"))

	return nil
}

func listDocuments(ctx context.Context, c *cli.Command) error {
	// データベース初期化
	database, err := db.Init(nil)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer db.Close(database)

	// ユースケース初期化
	documentUsecase := usecase.NewDocumentUsecase(database)

	// 全ドキュメント取得
	documents, err := documentUsecase.GetAllDocuments()
	if err != nil {
		return fmt.Errorf("failed to get documents: %w", err)
	}

	if len(documents) == 0 {
		fmt.Println("No documents found.")
		return nil
	}

	fmt.Printf("Found %d documents:\n\n", len(documents))
	for _, doc := range documents {
		fmt.Printf("ID: %d\n", doc.ID)
		fmt.Printf("Title: %s\n", doc.Title)
		fmt.Printf("Summary: %s\n", doc.Summary)
		fmt.Printf("Created: %s\n", doc.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Println("---")
	}

	return nil
}

func showDocument(ctx context.Context, c *cli.Command) error {
	id := c.Int("id")

	// データベース初期化
	database, err := db.Init(nil)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer db.Close(database)

	// ユースケース初期化
	documentUsecase := usecase.NewDocumentUsecase(database)

	// ドキュメント取得
	document, err := documentUsecase.GetDocument(uint(id))
	if err != nil {
		return fmt.Errorf("failed to get document: %w", err)
	}

	// 詳細表示
	fmt.Printf("=== Document ID: %d ===\n", document.ID)
	fmt.Printf("Title: %s\n", document.Title)
	fmt.Printf("Summary: %s\n", document.Summary)
	fmt.Printf("Created: %s\n", document.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("Updated: %s\n", document.UpdatedAt.Format("2006-01-02 15:04:05"))
	fmt.Println("\n=== Content ===")
	fmt.Println(document.Content)

	return nil
}

func createDocuments(ctx context.Context, c *cli.Command) error {
	fmt.Printf("Creating documents from fragments using AI...\n\n")

	// データベース初期化
	database, err := db.Init(nil)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer db.Close(database)

	// AIサービス初期化
	aiService, err := ai.NewService(database)
	if err != nil {
		return fmt.Errorf("failed to create AI service: %w", err)
	}

	// ドキュメント作成
	err = aiService.CreateDocuments(ctx)
	if err != nil {
		return fmt.Errorf("failed to create documents: %w", err)
	}

	return nil
}

func compressFragments(ctx context.Context, c *cli.Command) error {
	fmt.Printf("Compressing fragments using AI...\n\n")

	// データベース初期化
	database, err := db.Init(nil)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer db.Close(database)

	// AIサービス初期化
	aiService, err := ai.NewService(database)
	if err != nil {
		return fmt.Errorf("failed to create AI service: %w", err)
	}

	// フラグメント圧縮
	err = aiService.CompressFragments(ctx)
	if err != nil {
		return fmt.Errorf("failed to compress fragments: %w", err)
	}

	return nil
}

func listFragments(ctx context.Context, c *cli.Command) error {
	// データベース初期化
	database, err := db.Init(nil)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer db.Close(database)

	// フラグメント取得
	fragmentUsecase := usecase.NewFragmentUsecase(database)
	fragments, err := fragmentUsecase.GetAllFragments()
	if err != nil {
		return fmt.Errorf("failed to get fragments: %w", err)
	}

	if len(fragments) == 0 {
		fmt.Println("No fragments found.")
		return nil
	}

	fmt.Printf("Found %d fragments:\n\n", len(fragments))
	for _, fragment := range fragments {
		fmt.Printf("ID: %d\n", fragment.ID)
		fmt.Printf("Content: %s\n", fragment.Content)
		fmt.Printf("Created: %s\n", fragment.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Println("---")
	}

	return nil
}

func deleteFragment(ctx context.Context, c *cli.Command) error {
	id := c.Int("id")
	if id <= 0 {
		return fmt.Errorf("invalid fragment ID: %d", id)
	}

	// データベース初期化
	database, err := db.Init(nil)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer db.Close(database)

	fragmentUsecase := usecase.NewFragmentUsecase(database)

	// フラグメントが存在するかチェック
	fragment, err := fragmentUsecase.GetFragment(uint(id))
	if err != nil {
		return fmt.Errorf("fragment with ID %d not found", id)
	}

	// 削除実行
	err = fragmentUsecase.DeleteFragment(uint(id))
	if err != nil {
		return fmt.Errorf("failed to delete fragment: %w", err)
	}

	fmt.Printf("Fragment deleted successfully!\n")
	fmt.Printf("ID: %d\n", fragment.ID)
	fmt.Printf("Content: %s\n", fragment.Content)

	return nil
}
