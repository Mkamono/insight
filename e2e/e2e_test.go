package e2e

import (
	"bytes"
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3" // SQLiteドライバをインポート
)

const testDBFile = "test_knowledge.db" // テスト用データベースファイル名

// setupTestDB はテスト用のデータベースを初期化します。
func setupTestDB(t *testing.T) *sql.DB {
	// 既存のテストDBファイルを削除
	os.Remove(testDBFile)

	db, err := sql.Open("sqlite3", testDBFile)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// SQLスキーマファイルを読み込み、データベースを初期化
	sqlFiles := []string{
		"../sql/001_create_fragments.sql",
		"../sql/002_create_documents.sql",
		"../sql/003_create_document_fragment_links.sql",
		"../sql/004_create_tags.sql",
		"../sql/005_create_document_tag_links.sql",
	}

	for _, file := range sqlFiles {
		sqlContent, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("Failed to read SQL file %s: %v", file, err)
		}
		_, err = db.Exec(string(sqlContent))
		if err != nil {
			t.Fatalf("Failed to execute SQL from %s: %v", file, err)
		}
	}
	return db
}

// teardownTestDB はテスト用データベースをクリーンアップします。
func teardownTestDB(db *sql.DB) {
	db.Close()
	os.Remove(testDBFile)
}

// runInsightCommand はinsight CLIコマンドを実行し、その結果を返します。
func runInsightCommand(t *testing.T, args ...string) (string, string, int) {
	// プロジェクトルートのinsightバイナリへのパスを構築
	insightBinaryPath := filepath.Join("..", "insight")
	cmd := exec.Command(insightBinaryPath, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// テスト用データベースファイルを使用するように環境変数を設定
	cmd.Env = append(os.Environ(), fmt.Sprintf("INSIGHT_DB_FILE=%s", testDBFile))

	err := cmd.Run()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return stdout.String(), stderr.String(), exitError.ExitCode()
		}
		t.Fatalf("Failed to run command: %v, stderr: %s", err, stderr.String())
	}
	return stdout.String(), stderr.String(), 0
}

// Test cases for insight add command
func TestInsightAddCommand(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		expectedStdout string
		expectedStderr string
		expectedExit   int
		checkDB        func(*testing.T, *sql.DB) // データベースの状態をチェックする関数
	}{
		{
			name:           "add with text only",
			args:           []string{"add", "A simple text fragment."},
			expectedStdout: "Fragment added with ID: 1\n", // IDが1であることを期待
			expectedStderr: "",
			expectedExit:   0,
			checkDB: func(t *testing.T, db *sql.DB) {
				var content string
				err := db.QueryRow("SELECT content FROM fragments WHERE id = 1").Scan(&content)
				if err != nil {
					t.Fatalf("Failed to query fragment from DB: %v", err)
				}
				expectedContent := `{"text":"A simple text fragment."}`
				if content != expectedContent {
					t.Errorf("Expected fragment content: %q, Got: %q", expectedContent, content)
				}
			},
		},
		{
			name:           "add with no arguments",
			args:           []string{"add"},
			expectedStdout: "Usage: insight add [text] [--url <url>] [--image <path>]\n",
			expectedStderr: "",
			expectedExit:   1,   // 期待される終了コードは1 (エラー)
			checkDB:        nil, // データベースチェックは不要
		},
		{
			name:           "add with url only",
			args:           []string{"add", "--url", "https://example.com/article"},
			expectedStdout: "Fragment added with ID: 1\n",
			expectedStderr: "",
			expectedExit:   0,
			checkDB: func(t *testing.T, db *sql.DB) {
				var content string
				err := db.QueryRow("SELECT content FROM fragments WHERE id = 1").Scan(&content)
				if err != nil {
					t.Fatalf("Failed to query fragment from DB: %v", err)
				}
				expectedContent := `{"url":"https://example.com/article"}`
				if content != expectedContent {
					t.Errorf("Expected fragment content: %q, Got: %q", expectedContent, content)
				}
			},
		},
		{
			name:           "add with text and url (flags first)",
			args:           []string{"add", "--url", "https://example.com/article", "Interesting article."},
			expectedStdout: "Fragment added with ID: 1\n",
			expectedStderr: "",
			expectedExit:   0,
			checkDB: func(t *testing.T, db *sql.DB) {
				var content string
				err := db.QueryRow("SELECT content FROM fragments WHERE id = 1").Scan(&content)
				if err != nil {
					t.Fatalf("Failed to query fragment from DB: %v", err)
				}
				expectedContent := `{"text":"Interesting article.","url":"https://example.com/article"}`
				if content != expectedContent {
					t.Errorf("Expected fragment content: %q, Got: %q", expectedContent, content)
				}
			},
		},
		{
			name:           "add with image only",
			args:           []string{"add", "--image", "~/images/diagram.png"},
			expectedStdout: "Fragment added with ID: 1\n",
			expectedStderr: "",
			expectedExit:   0,
			checkDB: func(t *testing.T, db *sql.DB) {
				var content string
				err := db.QueryRow("SELECT content FROM fragments WHERE id = 1").Scan(&content)
				if err != nil {
					t.Fatalf("Failed to query fragment from DB: %v", err)
				}
				expectedContent := `{"image_path":"~/images/diagram.png"}`
				if content != expectedContent {
					t.Errorf("Expected fragment content: %q, Got: %q", expectedContent, content)
				}
			},
		},
		{
			name:           "add with text, url, and image",
			args:           []string{"add", "--url", "https://example.com/article", "--image", "path/to/figure.png", "This figure in the article is key."},
			expectedStdout: "Fragment added with ID: 1\n",
			expectedStderr: "",
			expectedExit:   0,
			checkDB: func(t *testing.T, db *sql.DB) {
				var content string
				err := db.QueryRow("SELECT content FROM fragments WHERE id = 1").Scan(&content)
				if err != nil {
					t.Fatalf("Failed to query fragment from DB: %v", err)
				}
				expectedContent := `{"image_path":"path/to/figure.png","text":"This figure in the article is key.","url":"https://example.com/article"}`
				if content != expectedContent {
					t.Errorf("Expected fragment content: %q, Got: %q", expectedContent, content)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)     // 各テストケースの前にDBを初期化
			defer teardownTestDB(db) // 各テストケースの後にDBをクリーンアップ

			stdout, stderr, exitCode := runInsightCommand(t, tt.args...)

			if stdout != tt.expectedStdout {
				t.Errorf("Expected stdout: %q, Got: %q", tt.expectedStdout, stdout)
			}
			// stderrのタイムスタンプが動的な場合は正規表現でチェック
			if tt.name == "doc show with non-existent document" {
				expectedStderrPattern := `\d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2} Failed to get document: document not found: 999\n`
				if !regexp.MustCompile(expectedStderrPattern).MatchString(stderr) {
					t.Errorf("Expected stderr to match pattern, Got: %q", stderr)
				}
			} else {
				if stderr != tt.expectedStderr {
					t.Errorf("Expected stderr: %q, Got: %q", tt.expectedStderr, stderr)
				}
			}
			if exitCode != tt.expectedExit {
				t.Errorf("Expected exit code: %d, Got: %d", tt.expectedExit, exitCode)
			}

			if tt.checkDB != nil {
				tt.checkDB(t, db)
			}
		})
	}
}

// Test cases for insight process command
func TestInsightProcessCommand(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		expectedStdout string
		expectedStderr string
		expectedExit   int
	}{
		{
			name:           "process without flags",
			args:           []string{"process"},
			expectedStdout: "Process command executed.\n",
			expectedStderr: "",
			expectedExit:   0,
		},
		{
			name:           "process with dry-run flag",
			args:           []string{"process", "--dry-run"},
			expectedStdout: "Process command executed.\nDry run mode enabled.\n",
			expectedStderr: "",
			expectedExit:   0,
		},
		{
			name:           "process with doc flag",
			args:           []string{"process", "--doc", "my_document_id"},
			expectedStdout: "Process command executed.\nProcessing document: my_document_id\n",
			expectedStderr: "",
			expectedExit:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)     // 各テストケースの前にDBを初期化
			defer teardownTestDB(db) // 各テストケースの後にDBをクリーンアップ

			stdout, stderr, exitCode := runInsightCommand(t, tt.args...)

			if stdout != tt.expectedStdout {
				t.Errorf("Expected stdout: %q, Got: %q", tt.expectedStdout, stdout)
			}
			// stderrのタイムスタンプが動的な場合は正規表現でチェック
			if tt.name == "doc show with non-existent document" {
				expectedStderrPattern := `\d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2} Failed to get document: document not found: 999\n`
				if !regexp.MustCompile(expectedStderrPattern).MatchString(stderr) {
					t.Errorf("Expected stderr to match pattern, Got: %q", stderr)
				}
			} else {
				if stderr != tt.expectedStderr {
					t.Errorf("Expected stderr: %q, Got: %q", tt.expectedStderr, stderr)
				}
			}
			if exitCode != tt.expectedExit {
				t.Errorf("Expected exit code: %d, Got: %d", tt.expectedExit, exitCode)
			}
		})
	}
}

// Test cases for insight doc list command
func TestInsightDocListCommand(t *testing.T) {
	tests := []struct {
		name           string
		setupDB        func(*testing.T, *sql.DB) // テスト前にDBにデータを挿入する関数
		args           []string
		expectedStdout string
		expectedStderr string
		expectedExit   int
	}{
		{
			name:           "doc list without documents",
			setupDB:        nil, // 何もしない
			args:           []string{"doc", "list"},
			expectedStdout: "No documents found.\n",
			expectedStderr: "",
			expectedExit:   0,
		},
		{
			name: "doc list with documents",
			setupDB: func(t *testing.T, db *sql.DB) {
				// テスト用のドキュメントを挿入
				_, err := db.Exec(`INSERT INTO documents (id, title, content_markdown, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`,
					1, "Test Document 1", "Content 1", time.Now().Add(-24*time.Hour).Format("2006-01-02 15:04:05"), time.Now().Add(-24*time.Hour).Format("2006-01-02 15:04:05"))
				if err != nil {
					t.Fatalf("Failed to insert test document: %v", err)
				}
				_, err = db.Exec(`INSERT INTO documents (id, title, content_markdown, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`,
					2, "Another Document", "Content 2", time.Now().Format("2006-01-02 15:04:05"), time.Now().Format("2006-01-02 15:04:05"))
				if err != nil {
					t.Fatalf("Failed to insert test document: %v", err)
				}
			},
			args:           []string{"doc", "list"},
			expectedStdout: "ID\tTitle\tLast Updated\n--\t-----\t------------\n2\tAnother Document\t" + time.Now().Format("2006-01-02 15:04:05") + "\n1\tTest Document 1\t" + time.Now().Add(-24*time.Hour).Format("2006-01-02 15:04:05") + "\n",
			expectedStderr: "",
			expectedExit:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)     // 各テストケースの前にDBを初期化
			defer teardownTestDB(db) // 各テストケースの後にDBをクリーンアップ

			if tt.setupDB != nil {
				tt.setupDB(t, db) // テスト固有のDBセットアップ
			}

			stdout, stderr, exitCode := runInsightCommand(t, tt.args...)

			// 日付部分は動的に生成されるため、正規表現でマッチング
			if tt.name == "doc list with documents" {
				expectedOutputPattern := `ID\tTitle\tLast Updated\n--\t-----\t------------\n2\tAnother Document\t\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\n1\tTest Document 1\t\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\n`
				if !regexp.MustCompile(expectedOutputPattern).MatchString(stdout) {
					t.Errorf("Expected stdout to match pattern: %q, Got: %q", expectedOutputPattern, stdout)
				}
			} else {
				if stdout != tt.expectedStdout {
					t.Errorf("Expected stdout: %q, Got: %q", tt.expectedStdout, stdout)
				}
			}

			if stderr != tt.expectedStderr {
				t.Errorf("Expected stderr: %q, Got: %q", tt.expectedStderr, stderr)
			}
			if exitCode != tt.expectedExit {
				t.Errorf("Expected exit code: %d, Got: %d", tt.expectedExit, exitCode)
			}
		})
	}
}

// Test cases for insight doc show command
func TestInsightDocShowCommand(t *testing.T) {
	tests := []struct {
		name           string
		setupDB        func(*testing.T, *sql.DB)
		args           []string
		expectedStdout string
		expectedStderr string
		expectedExit   int
	}{
		{
			name:           "doc show without arguments",
			setupDB:        nil,
			args:           []string{"doc", "show"},
			expectedStdout: "Usage: insight doc show <id_or_title>\n",
			expectedStderr: "",
			expectedExit:   1,
		},
		{
			name: "doc show with existing document by ID",
			setupDB: func(t *testing.T, db *sql.DB) {
				_, err := db.Exec(`INSERT INTO documents (id, title, content_markdown, summary, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`,
					1, "Test Document", "# Test Content\n\nThis is a test document.", "Test summary",
					time.Now().Add(-24*time.Hour).Format("2006-01-02 15:04:05"), time.Now().Format("2006-01-02 15:04:05"))
				if err != nil {
					t.Fatalf("Failed to insert test document: %v", err)
				}
			},
			args:           []string{"doc", "show", "1"},
			expectedStdout: "", // We'll check this with regex since it contains dynamic timestamps
			expectedStderr: "",
			expectedExit:   0,
		},
		{
			name:           "doc show with non-existent document",
			setupDB:        nil,
			args:           []string{"doc", "show", "999"},
			expectedStdout: "",
			expectedStderr: "2025/07/03 14:35:58 Failed to get document: document not found: 999\n", // log.Fatalfの出力を期待
			expectedExit:   1,                                                                       // log.Fatalf will cause exit code 1
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)     // 各テストケースの前にDBを初期化
			defer teardownTestDB(db) // 各テストケースの後にDBをクリーンアップ

			if tt.setupDB != nil {
				tt.setupDB(t, db)
			}

			stdout, stderr, exitCode := runInsightCommand(t, tt.args...)

			if tt.name == "doc show with existing document by ID" {
				// 動的なタイムスタンプが含まれるため、パターンマッチングで確認
				if !regexp.MustCompile(`# Test Document`).MatchString(stdout) {
					t.Errorf("Expected stdout to contain '# Test Document', Got: %q", stdout)
				}
				if !regexp.MustCompile(`\*\*ID:\*\* 1`).MatchString(stdout) {
					t.Errorf("Expected stdout to contain '**ID:** 1', Got: %q", stdout)
				}
				if !regexp.MustCompile(`## Content`).MatchString(stdout) {
					t.Errorf("Expected stdout to contain '## Content', Got: %q", stdout)
				}
			} else {
				if stdout != tt.expectedStdout {
					t.Errorf("Expected stdout: %q, Got: %q", tt.expectedStdout, stdout)
				}
			}

			// stderrのタイムスタンプが動的な場合は正規表現でチェック
			if tt.name == "doc show with non-existent document" {
				expectedStderrPattern := `\d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2} Failed to get document: document not found: 999\n`
				if !regexp.MustCompile(expectedStderrPattern).MatchString(stderr) {
					t.Errorf("Expected stderr to match pattern, Got: %q", stderr)
				}
			} else {
				if stderr != tt.expectedStderr {
					t.Errorf("Expected stderr: %q, Got: %q", tt.expectedStderr, stderr)
				}
			}
			if exitCode != tt.expectedExit {
				t.Errorf("Expected exit code: %d, Got: %d", tt.expectedExit, exitCode)
			}
		})
	}
}

// Test cases for insight tag list command
func TestInsightTagListCommand(t *testing.T) {
	tests := []struct {
		name           string
		setupDB        func(*testing.T, *sql.DB)
		args           []string
		expectedStdout string
		expectedStderr string
		expectedExit   int
	}{
		{
			name:           "tag list without tags",
			setupDB:        nil,
			args:           []string{"tag", "list"},
			expectedStdout: "No tags found.\n",
			expectedStderr: "",
			expectedExit:   0,
		},
		{
			name: "tag list with tags",
			setupDB: func(t *testing.T, db *sql.DB) {
				// テスト用のタグとドキュメントを挿入
				_, err := db.Exec(`INSERT INTO tags (id, name) VALUES (?, ?)`, 1, "Go")
				if err != nil {
					t.Fatalf("Failed to insert tag: %v", err)
				}
				_, err = db.Exec(`INSERT INTO tags (id, name) VALUES (?, ?)`, 2, "Testing")
				if err != nil {
					t.Fatalf("Failed to insert tag: %v", err)
				}
				_, err = db.Exec(`INSERT INTO documents (id, title, content_markdown, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`,
					1, "Go E2E Testing", "Content", time.Now().Format("2006-01-02 15:04:05"), time.Now().Format("2006-01-02 15:04:05"))
				if err != nil {
					t.Fatalf("Failed to insert document: %v", err)
				}
				_, err = db.Exec(`INSERT INTO documents (id, title, content_markdown, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`,
					2, "Go Concurrency", "Content", time.Now().Format("2006-01-02 15:04:05"), time.Now().Format("2006-01-02 15:04:05"))
				if err != nil {
					t.Fatalf("Failed to insert document: %v", err)
				}
				_, err = db.Exec(`INSERT INTO document_tag_links (document_id, tag_id) VALUES (?, ?)`, 1, 1) // Go E2E Testing -> Go
				if err != nil {
					t.Fatalf("Failed to insert link: %v", err)
				}
				_, err = db.Exec(`INSERT INTO document_tag_links (document_id, tag_id) VALUES (?, ?)`, 1, 2) // Go E2E Testing -> Testing
				if err != nil {
					t.Fatalf("Failed to insert link: %v", err)
				}
				_, err = db.Exec(`INSERT INTO document_tag_links (document_id, tag_id) VALUES (?, ?)`, 2, 1) // Go Concurrency -> Go
				if err != nil {
					t.Fatalf("Failed to insert link: %v", err)
				}
			},
			args:           []string{"tag", "list"},
			expectedStdout: "ID\tTag\tCount\n--\t-----\t-----\n1\tGo\t2\n2\tTesting\t1\n",
			expectedStderr: "",
			expectedExit:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)     // 各テストケースの前にDBを初期化
			defer teardownTestDB(db) // 各テストケースの後にDBをクリーンアップ

			if tt.setupDB != nil {
				tt.setupDB(t, db) // テスト固有のDBセットアップ
			}

			stdout, stderr, exitCode := runInsightCommand(t, tt.args...)

			if stdout != tt.expectedStdout {
				t.Errorf("Expected stdout: %q, Got: %q", tt.expectedStdout, stdout)
			}
			if stderr != tt.expectedStderr {
				t.Errorf("Expected stderr: %q, Got: %q", tt.expectedStderr, stderr)
			}
			if exitCode != tt.expectedExit {
				t.Errorf("Expected exit code: %d, Got: %d", tt.expectedExit, exitCode)
			}
		})
	}
}

// Test cases for insight find tag command
func TestInsightFindTagCommand(t *testing.T) {
	tests := []struct {
		name           string
		setupDB        func(*testing.T, *sql.DB)
		args           []string
		expectedStdout string
		expectedStderr string
		expectedExit   int
	}{
		{
			name:           "find without arguments",
			setupDB:        nil,
			args:           []string{"find"},
			expectedStdout: "Usage: insight find --tag <tag_name> | --query \"<search_term>\"\n",
			expectedStderr: "",
			expectedExit:   1,
		},
		{
			name:           "find with tag but no tag name",
			setupDB:        nil,
			args:           []string{"find", "--tag"},
			expectedStdout: "",                                                                                                                                                 // flagパッケージの挙動により、エラーメッセージはstderrに出力される
			expectedStderr: "flag needs an argument: -tag\nUsage of find:\n  -query string\n    \tSearch query for full-text search\n  -tag string\n    \tTag to search for\n", // 修正
			expectedExit:   2,                                                                                                                                                  // 修正
		},
		{
			name: "find with existing tag",
			setupDB: func(t *testing.T, db *sql.DB) {
				// テストデータを挿入
				_, err := db.Exec(`INSERT INTO documents (id, title, content_markdown) VALUES (?, ?, ?)`,
					1, "Go Document", "Content about Go")
				if err != nil {
					t.Fatalf("Failed to insert document: %v", err)
				}
				_, err = db.Exec(`INSERT INTO documents (id, title, content_markdown) VALUES (?, ?, ?)`,
					2, "Python Document", "Content about Python")
				if err != nil {
					t.Fatalf("Failed to insert document: %v", err)
				}
				_, err = db.Exec(`INSERT INTO tags (id, name) VALUES (?, ?)`, 1, "Go")
				if err != nil {
					t.Fatalf("Failed to insert tag: %v", err)
				}
				_, err = db.Exec(`INSERT INTO document_tag_links (document_id, tag_id) VALUES (?, ?)`, 1, 1)
				if err != nil {
					t.Fatalf("Failed to link document and tag: %v", err)
				}
			},
			args:           []string{"find", "--tag", "Go"},
			expectedStdout: "Found 1 document(s) with tag 'Go':\n\nID\tTitle\tLast Updated\n--\t-----\t------------\n1\tGo Document\t" + time.Now().Format("2006-01-02 15:04:05") + "\n",
			expectedStderr: "",
			expectedExit:   0,
		},
		{
			name:           "find with non-existent tag",
			setupDB:        nil,
			args:           []string{"find", "--tag", "NonExistent"},
			expectedStdout: "No documents found with tag: NonExistent\n",
			expectedStderr: "",
			expectedExit:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)     // 各テストケースの前にDBを初期化
			defer teardownTestDB(db) // 各テストケースの後にDBをクリーンアップ

			if tt.setupDB != nil {
				tt.setupDB(t, db)
			}

			stdout, stderr, exitCode := runInsightCommand(t, tt.args...)

			// 動的なタイムスタンプが含まれる場合は正規表現でチェック
			if tt.name == "find with existing tag" {
				expectedPattern := `Found 1 document\(s\) with tag 'Go':\n\nID\tTitle\tLast Updated\n--\t-----\t------------\n1\tGo Document\t\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\n`
				if !regexp.MustCompile(expectedPattern).MatchString(stdout) {
					t.Errorf("Expected stdout to match pattern, Got: %q", stdout)
				}
			} else {
				if stdout != tt.expectedStdout {
					t.Errorf("Expected stdout: %q, Got: %q", tt.expectedStdout, stdout)
				}
			}

			if stderr != tt.expectedStderr {
				t.Errorf("Expected stderr: %q, Got: %q", tt.expectedStderr, stderr)
			}
			if exitCode != tt.expectedExit {
				t.Errorf("Expected exit code: %d, Got: %d", tt.expectedExit, exitCode)
			}
		})
	}
}

// Test cases for insight find query command
func TestInsightFindQueryCommand(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		expectedStdout string
		expectedStderr string
		expectedExit   int
	}{
		{
			name:           "find query without query string",
			args:           []string{"find", "--query"},
			expectedStdout: "",
			expectedStderr: "flag needs an argument: -query\nUsage of find:\n  -query string\n    \tSearch query for full-text search\n  -tag string\n    \tTag to search for\n",
			expectedExit:   2,
		},
		{
			name:           "find query with search term",
			args:           []string{"find", "--query", "search term"},
			expectedStdout: "No documents found matching query: search term\n",
			expectedStderr: "",
			expectedExit:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)     // 各テストケースの前にDBを初期化
			defer teardownTestDB(db) // 各テストケースの後にDBをクリーンアップ

			stdout, stderr, exitCode := runInsightCommand(t, tt.args...)

			if stdout != tt.expectedStdout {
				t.Errorf("Expected stdout: %q, Got: %q", tt.expectedStdout, stdout)
			}
			if stderr != tt.expectedStderr {
				t.Errorf("Expected stderr: %q, Got: %q", tt.expectedStderr, stderr)
			}
			if exitCode != tt.expectedExit {
				t.Errorf("Expected exit code: %d, Got: %d", tt.expectedExit, exitCode)
			}
		})
	}
}

// Test cases for insight status command
func TestInsightStatusCommand(t *testing.T) {
	tests := []struct {
		name           string
		setupDB        func(*testing.T, *sql.DB)
		args           []string
		expectedStdout string
		expectedStderr string
		expectedExit   int
	}{
		{
			name:           "status command - empty db",
			setupDB:        nil,
			args:           []string{"status"},
			expectedStdout: "Documents: 0\nTotal Fragments: 0\nUnprocessed Fragments: 0\n",
			expectedStderr: "",
			expectedExit:   0,
		},
		{
			name: "status command - with fragments",
			setupDB: func(t *testing.T, db *sql.DB) {
				_, err := db.Exec(`INSERT INTO fragments (content) VALUES (?)`, `{"text":"fragment 1"}`)
				if err != nil {
					t.Fatalf("Failed to insert fragment: %v", err)
				}
				_, err = db.Exec(`INSERT INTO fragments (content) VALUES (?)`, `{"text":"fragment 2"}`)
				if err != nil {
					t.Fatalf("Failed to insert fragment: %v", err)
				}
			},
			args:           []string{"status"},
			expectedStdout: "Documents: 0\nTotal Fragments: 2\nUnprocessed Fragments: 2\n",
			expectedStderr: "",
			expectedExit:   0,
		},
		{
			name: "status command - with documents and processed fragments",
			setupDB: func(t *testing.T, db *sql.DB) {
				// Insert fragments
				_, err := db.Exec(`INSERT INTO fragments (id, content) VALUES (?, ?)`, 1, `{"text":"fragment 1"}`)
				if err != nil {
					t.Fatalf("Failed to insert fragment: %v", err)
				}
				_, err = db.Exec(`INSERT INTO fragments (id, content) VALUES (?, ?)`, 2, `{"text":"fragment 2"}`)
				if err != nil {
					t.Fatalf("Failed to insert fragment: %v", err)
				}
				_, err = db.Exec(`INSERT INTO fragments (id, content) VALUES (?, ?)`, 3, `{"text":"fragment 3"}`)
				if err != nil {
					t.Fatalf("Failed to insert fragment: %v", err)
				}

				// Insert documents
				_, err = db.Exec(`INSERT INTO documents (id, title, content_markdown) VALUES (?, ?, ?)`, 1, "Doc 1", "Content 1")
				if err != nil {
					t.Fatalf("Failed to insert document: %v", err)
				}
				_, err = db.Exec(`INSERT INTO documents (id, title, content_markdown) VALUES (?, ?, ?)`, 2, "Doc 2", "Content 2")
				if err != nil {
					t.Fatalf("Failed to insert document: %v", err)
				}

				// Link fragments to documents (process some fragments)
				_, err = db.Exec(`INSERT INTO document_fragment_links (document_id, fragment_id) VALUES (?, ?)`, 1, 1)
				if err != nil {
					t.Fatalf("Failed to link fragment: %v", err)
				}
				_, err = db.Exec(`INSERT INTO document_fragment_links (document_id, fragment_id) VALUES (?, ?)`, 1, 2)
				if err != nil {
					t.Fatalf("Failed to link fragment: %v", err)
				}
			},
			args:           []string{"status"},
			expectedStdout: "Documents: 2\nTotal Fragments: 3\nUnprocessed Fragments: 1\n",
			expectedStderr: "",
			expectedExit:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)     // 各テストケースの前にDBを初期化
			defer teardownTestDB(db) // 各テストケースの後にDBをクリーンアップ

			if tt.setupDB != nil {
				tt.setupDB(t, db) // テスト固有のDBセットアップ
			}

			stdout, stderr, exitCode := runInsightCommand(t, tt.args...)

			if stdout != tt.expectedStdout {
				t.Errorf("Expected stdout: %q, Got: %q", tt.expectedStdout, stdout)
			}
			if stderr != tt.expectedStderr {
				t.Errorf("Expected stderr: %q, Got: %q", tt.expectedStderr, stderr)
			}
			if exitCode != tt.expectedExit {
				t.Errorf("Expected exit code: %d, Got: %d", tt.expectedExit, exitCode)
			}
		})
	}
}

// Test cases for insight fragment list command
func TestInsightFragmentListCommand(t *testing.T) {
	tests := []struct {
		name           string
		setupDB        func(*testing.T, *sql.DB)
		args           []string
		expectedStdout string
		expectedStderr string
		expectedExit   int
	}{
		{
			name:           "fragment list - empty db",
			setupDB:        nil,
			args:           []string{"fragment", "list"},
			expectedStdout: "No fragments found.\n",
			expectedStderr: "",
			expectedExit:   0,
		},
		{
			name: "fragment list - with fragments",
			setupDB: func(t *testing.T, db *sql.DB) {
				_, err := db.Exec(`INSERT INTO fragments (id, content, created_at) VALUES (?, ?, ?)`, 1, `{"text":"fragment 1"}`, time.Now().Add(-24*time.Hour).Format("2006-01-02 15:04:05"))
				if err != nil {
					t.Fatalf("Failed to insert fragment: %v", err)
				}
				_, err = db.Exec(`INSERT INTO fragments (id, content, created_at) VALUES (?, ?, ?)`, 2, `{"text":"fragment 2"}`, time.Now().Format("2006-01-02 15:04:05"))
				if err != nil {
					t.Fatalf("Failed to insert fragment: %v", err)
				}
			},
			args:           []string{"fragment", "list"},
			expectedStdout: "ID\tContent\tCreated At\n--\t-------\t----------\n2\t{\"text\":\"fragment 2\"}\t" + time.Now().Format("2006-01-02 15:04:05") + "\n1\t{\"text\":\"fragment 1\"}\t" + time.Now().Add(-24*time.Hour).Format("2006-01-02 15:04:05") + "\n",
			expectedStderr: "",
			expectedExit:   0,
		},
		{
			name: "fragment list - with unprocessed fragments",
			setupDB: func(t *testing.T, db *sql.DB) {
				// Insert fragments
				_, err := db.Exec(`INSERT INTO fragments (id, content) VALUES (?, ?)`, 1, `{"text":"fragment 1"}`)
				if err != nil {
					t.Fatalf("Failed to insert fragment: %v", err)
				}
				_, err = db.Exec(`INSERT INTO fragments (id, content) VALUES (?, ?)`, 2, `{"text":"fragment 2"}`)
				if err != nil {
					t.Fatalf("Failed to insert fragment: %v", err)
				}
				_, err = db.Exec(`INSERT INTO fragments (id, content) VALUES (?, ?)`, 3, `{"text":"fragment 3"}`)
				if err != nil {
					t.Fatalf("Failed to insert fragment: %v", err)
				}

				// Link fragment 1 and 2 to a document (process them)
				_, err = db.Exec(`INSERT INTO documents (id, title, content_markdown) VALUES (?, ?, ?)`, 1, "Doc 1", "Content 1")
				if err != nil {
					t.Fatalf("Failed to insert document: %v", err)
				}
				_, err = db.Exec(`INSERT INTO document_fragment_links (document_id, fragment_id) VALUES (?, ?)`, 1, 1)
				if err != nil {
					t.Fatalf("Failed to link fragment: %v", err)
				}
				_, err = db.Exec(`INSERT INTO document_fragment_links (document_id, fragment_id) VALUES (?, ?)`, 1, 2)
				if err != nil {
					t.Fatalf("Failed to link fragment: %v", err)
				}
			},
			args:           []string{"fragment", "list", "--unprocessed"},
			expectedStdout: "ID\tContent\tCreated At\n--\t-------\t----------\n3\t{\"text\":\"fragment 3\"}\t" + time.Now().Format("2006-01-02 15:04:05") + "\n",
			expectedStderr: "",
			expectedExit:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)     // 各テストケースの前にDBを初期化
			defer teardownTestDB(db) // 各テストケースの後にDBをクリーンアップ

			if tt.setupDB != nil {
				tt.setupDB(t, db) // テスト固有のDBセットアップ
			}

			stdout, stderr, exitCode := runInsightCommand(t, tt.args...)

			// 日付部分は動的に生成されるため、正規表現でマッチング
			if tt.name == "fragment list - with fragments" {
				expectedOutputPattern := `ID\tContent\tCreated At\n--\t-------\t----------\n2\t{\"text\":\"fragment 2\"}\t\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\n1\t{\"text\":\"fragment 1\"}\t\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\n`
				if !regexp.MustCompile(expectedOutputPattern).MatchString(stdout) {
					t.Errorf("Expected stdout to match pattern: %q, Got: %q", expectedOutputPattern, stdout)
				}
			} else if tt.name == "fragment list - with unprocessed fragments" {
				expectedOutputPattern := `ID\tContent\tCreated At\n--\t-------\t----------\n3\t{\"text\":\"fragment 3\"}\t\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\n`
				if !regexp.MustCompile(expectedOutputPattern).MatchString(stdout) {
					t.Errorf("Expected stdout to match pattern: %q, Got: %q", expectedOutputPattern, stdout)
				}
			} else {
				if stdout != tt.expectedStdout {
					t.Errorf("Expected stdout: %q, Got: %q", tt.expectedStdout, stdout)
				}
			}

			if stderr != tt.expectedStderr {
				t.Errorf("Expected stderr: %q, Got: %q", tt.expectedStderr, stderr)
			}
			if exitCode != tt.expectedExit {
				t.Errorf("Expected exit code: %d, Got: %d", tt.expectedExit, exitCode)
			}
		})
	}
}

// TestSQLiteDriver はSQLiteドライバが正しく登録されているかを確認します。
func TestSQLiteDriver(t *testing.T) {
	dbFile := "test_driver.db"
	os.Remove(dbFile) // 以前のテストファイルをクリーンアップ

	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		t.Fatalf("Failed to open database with sqlite3 driver: %v", err)
	}
	defer db.Close()
	defer os.Remove(dbFile)

	err = db.Ping()
	if err != nil {
		t.Fatalf("Failed to ping database: %v", err)
	}
	t.Log("SQLite driver successfully opened and pinged database.")
}
