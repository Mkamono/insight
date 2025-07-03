package main

import (
	"insight/internal/cli"
	"insight/internal/database"
	"log"
	"os"
)

func main() {
	// データベースの初期化
	db := database.InitDB()
	defer db.Close()
	// CLIアプリケーションを起動
	app := cli.NewApp(db)
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
