package main

import (
	"fmt"
	"log"

	"insight/src/db"
	"insight/src/models"
)

func main() {
	fmt.Println("Starting database migration...")

	// データベース初期化
	database, err := db.Init(nil) // デフォルト設定を使用
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close(database)

	fmt.Println("Connected to SQLite database")

	// マイグレーション実行
	err = db.Migrate(database)
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	// マイグレーション完了をログ出力
	modelNames := models.GetModelNames()
	for _, name := range modelNames {
		fmt.Printf("✓ %s table migrated\n", name)
	}

	// データベース接続情報を表示
	sqlDB, err := database.DB()
	if err != nil {
		log.Printf("Warning: failed to get database stats: %v", err)
	} else {
		stats := sqlDB.Stats()
		fmt.Printf("Database connection stats - Open: %d, InUse: %d\n",
			stats.OpenConnections, stats.InUse)
	}

	fmt.Println("Migration completed successfully!")
}
