package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3" // SQLiteドライバをインポート
)

// Document はdocumentsテーブルの行を表す構造体です。
type Document struct {
	ID        int
	Title     string
	Content   string
	Summary   sql.NullString // NULLを許容するためsql.NullStringを使用
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Tag はtagsテーブルの行と関連ドキュメント数を表す構造体です。
type Tag struct {
	ID    int
	Name  string
	Count int
}

// Fragment はfragmentsテーブルの行を表す構造体です。
type Fragment struct {
	ID        int
	Content   string
	CreatedAt time.Time
}

// Question はquestionsテーブルの行を表す構造体です。
type Question struct {
	ID                 int
	QuestionText       string
	ContextFragmentIDs sql.NullString // JSON形式のフラグメントIDリスト
	ContextDocumentIDs sql.NullString // JSON形式のドキュメントIDリスト
	Status             string         // 'pending', 'answered', 'archived'
	CreatedAt          time.Time
	AnsweredAt         sql.NullTime
	AnswerFragmentID   sql.NullInt64
}

// InitDB はデータベースを初期化し、テーブルが存在しない場合は作成します。
func InitDB() *sql.DB {
	dbPath := os.Getenv("INSIGHT_DB_FILE")
	if dbPath == "" {
		dbPath = "knowledge.db" // デフォルトのデータベースファイル名
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	// fragmentsテーブルの作成
	createFragmentsTableSQL := `
CREATE TABLE IF NOT EXISTS fragments (
id INTEGER PRIMARY KEY AUTOINCREMENT,
content TEXT NOT NULL,
created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);`

	_, err = db.Exec(createFragmentsTableSQL)
	if err != nil {
		log.Fatalf("Failed to create fragments table: %v", err)
	}

	// documentsテーブルの作成
	createDocumentsTableSQL := `
CREATE TABLE IF NOT EXISTS documents (
id INTEGER PRIMARY KEY AUTOINCREMENT,
title TEXT NOT NULL,
content_markdown TEXT NOT NULL,
summary TEXT,
created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);`
	_, err = db.Exec(createDocumentsTableSQL)
	if err != nil {
		log.Fatalf("Failed to create documents table: %v", err)
	}

	// document_fragment_linksテーブルの作成
	createDocumentFragmentLinksTableSQL := `
CREATE TABLE IF NOT EXISTS document_fragment_links (
document_id INTEGER NOT NULL,
fragment_id INTEGER NOT NULL,
PRIMARY KEY (document_id, fragment_id),
FOREIGN KEY (document_id) REFERENCES documents(id),
FOREIGN KEY (fragment_id) REFERENCES fragments(id)
);`
	_, err = db.Exec(createDocumentFragmentLinksTableSQL)
	if err != nil {
		log.Fatalf("Failed to create document_fragment_links table: %v", err)
	}

	// tagsテーブルの作成
	createTagsTableSQL := `
CREATE TABLE IF NOT EXISTS tags (
id INTEGER PRIMARY KEY AUTOINCREMENT,
name TEXT NOT NULL UNIQUE
);`
	_, err = db.Exec(createTagsTableSQL)
	if err != nil {
		log.Fatalf("Failed to create tags table: %v", err)
	}

	// document_tag_linksテーブルの作成
	createDocumentTagLinksTableSQL := `
CREATE TABLE IF NOT EXISTS document_tag_links (
document_id INTEGER NOT NULL,
tag_id INTEGER NOT NULL,
PRIMARY KEY (document_id, tag_id),
FOREIGN KEY (document_id) REFERENCES documents(id),
FOREIGN KEY (tag_id) REFERENCES tags(id)
);`
	_, err = db.Exec(createDocumentTagLinksTableSQL)
	if err != nil {
		log.Fatalf("Failed to create document_tag_links table: %v", err)
	}

	// questionsテーブルの作成
	createQuestionsTableSQL := `
CREATE TABLE IF NOT EXISTS questions (
id INTEGER PRIMARY KEY AUTOINCREMENT,
question_text TEXT NOT NULL,
context_fragment_ids TEXT,
context_document_ids TEXT,
status TEXT DEFAULT 'pending',
created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
answered_at DATETIME,
answer_fragment_id INTEGER,
FOREIGN KEY (answer_fragment_id) REFERENCES fragments (id)
);`
	_, err = db.Exec(createQuestionsTableSQL)
	if err != nil {
		log.Fatalf("Failed to create questions table: %v", err)
	}

	return db
}

// InsertFragment は新しいフラグメントをデータベースに挿入します。
func InsertFragment(db *sql.DB, content string) (int64, error) {
	stmt, err := db.Prepare("INSERT INTO fragments(content, created_at) VALUES(?, ?)")
	if err != nil {
		return 0, fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(content, time.Now().Format(time.RFC3339))
	if err != nil {
		return 0, fmt.Errorf("failed to execute statement: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert ID: %w", err)
	}
	return id, nil
}

// GetAllDocuments はデータベースからすべてのドキュメントを取得します。
func GetAllDocuments(db *sql.DB) ([]Document, error) {
	rows, err := db.Query("SELECT id, title, content_markdown, summary, created_at, updated_at FROM documents ORDER BY updated_at DESC")
	if err != nil {
		return nil, fmt.Errorf("failed to query documents: %w", err)
	}
	defer rows.Close()

	var documents []Document
	for rows.Next() {
		var doc Document
		err := rows.Scan(&doc.ID, &doc.Title, &doc.Content, &doc.Summary, &doc.CreatedAt, &doc.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan document row: %w", err)
		}
		documents = append(documents, doc)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return documents, nil
}

// GetAllTags はデータベースからすべてのタグと関連ドキュメント数を取得します。
func GetAllTags(db *sql.DB) ([]Tag, error) {
	rows, err := db.Query(`
SELECT
t.id,
t.name,
COUNT(dtl.document_id) AS document_count
FROM
tags AS t
LEFT JOIN
document_tag_links AS dtl ON t.id = dtl.tag_id
GROUP BY
t.id, t.name
ORDER BY
t.name ASC
`)
	if err != nil {
		return nil, fmt.Errorf("failed to query tags: %w", err)
	}
	defer rows.Close()

	var tags []Tag
	for rows.Next() {
		var tag Tag
		err := rows.Scan(&tag.ID, &tag.Name, &tag.Count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tag row: %w", err)
		}
		tags = append(tags, tag)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return tags, nil
}

// GetCounts はドキュメント、フラグメント、未処理フラグメントの数を取得します。
func GetCounts(db *sql.DB) (int, int, int, error) {
	var docCount int
	err := db.QueryRow("SELECT COUNT(*) FROM documents").Scan(&docCount)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to get document count: %w", err)
	}

	var totalFragmentCount int
	err = db.QueryRow("SELECT COUNT(*) FROM fragments").Scan(&totalFragmentCount)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to get total fragment count: %w", err)
	}

	var unprocessedFragmentCount int
	err = db.QueryRow(`
SELECT COUNT(*)
FROM fragments f
LEFT JOIN document_fragment_links dfl ON f.id = dfl.fragment_id
WHERE dfl.fragment_id IS NULL
`).Scan(&unprocessedFragmentCount)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to get unprocessed fragment count: %w", err)
	}

	return docCount, totalFragmentCount, unprocessedFragmentCount, nil
}

// GetAllFragments はデータベースからすべてのフラグメントを取得します。
func GetAllFragments(db *sql.DB, unprocessed bool) ([]Fragment, error) {
	var rows *sql.Rows
	var err error

	if unprocessed {
		rows, err = db.Query(`
SELECT f.id, f.content, f.created_at
FROM fragments f
LEFT JOIN document_fragment_links dfl ON f.id = dfl.fragment_id
WHERE dfl.fragment_id IS NULL
ORDER BY f.created_at DESC
`)
	} else {
		rows, err = db.Query("SELECT id, content, created_at FROM fragments ORDER BY created_at DESC")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to query fragments: %w", err)
	}
	defer rows.Close()

	var fragments []Fragment
	for rows.Next() {
		var fragment Fragment
		err := rows.Scan(&fragment.ID, &fragment.Content, &fragment.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan fragment row: %w", err)
		}
		fragments = append(fragments, fragment)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return fragments, nil
}

// GetDocumentByIDOrTitle はIDまたはタイトルでドキュメントを取得します。
func GetDocumentByIDOrTitle(db *sql.DB, idOrTitle string) (*Document, error) {
	var doc Document
	var err error

	// まずIDとして扱って数値変換を試行
	if id, parseErr := strconv.Atoi(idOrTitle); parseErr == nil {
		err = db.QueryRow("SELECT id, title, content_markdown, summary, created_at, updated_at FROM documents WHERE id = ?", id).
			Scan(&doc.ID, &doc.Title, &doc.Content, &doc.Summary, &doc.CreatedAt, &doc.UpdatedAt)
	} else {
		// IDとして解析できない場合はタイトルとして検索
		err = db.QueryRow("SELECT id, title, content_markdown, summary, created_at, updated_at FROM documents WHERE title = ?", idOrTitle).
			Scan(&doc.ID, &doc.Title, &doc.Content, &doc.Summary, &doc.CreatedAt, &doc.UpdatedAt)
	}

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("document not found: %s", idOrTitle)
		}
		return nil, fmt.Errorf("failed to query document: %w", err)
	}

	return &doc, nil
}

// GetDocumentsByTag はタグでドキュメントを検索します。
func GetDocumentsByTag(db *sql.DB, tagName string) ([]Document, error) {
	rows, err := db.Query(`
SELECT d.id, d.title, d.content_markdown, d.summary, d.created_at, d.updated_at
FROM documents d
JOIN document_tag_links dtl ON d.id = dtl.document_id
JOIN tags t ON dtl.tag_id = t.id
WHERE t.name = ?
ORDER BY d.updated_at DESC
`, tagName)
	if err != nil {
		return nil, fmt.Errorf("failed to query documents by tag: %w", err)
	}
	defer rows.Close()

	var documents []Document
	for rows.Next() {
		var doc Document
		err := rows.Scan(&doc.ID, &doc.Title, &doc.Content, &doc.Summary, &doc.CreatedAt, &doc.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan document row: %w", err)
		}
		documents = append(documents, doc)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return documents, nil
}

// SearchDocuments は全文検索でドキュメントを検索します。
func SearchDocuments(db *sql.DB, searchTerm string) ([]Document, error) {
	searchPattern := "%" + searchTerm + "%"
	rows, err := db.Query(`
SELECT id, title, content_markdown, summary, created_at, updated_at
FROM documents
WHERE title LIKE ? OR content_markdown LIKE ? OR summary LIKE ?
ORDER BY updated_at DESC
`, searchPattern, searchPattern, searchPattern)
	if err != nil {
		return nil, fmt.Errorf("failed to search documents: %w", err)
	}
	defer rows.Close()

	var documents []Document
	for rows.Next() {
		var doc Document
		err := rows.Scan(&doc.ID, &doc.Title, &doc.Content, &doc.Summary, &doc.CreatedAt, &doc.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan document row: %w", err)
		}
		documents = append(documents, doc)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return documents, nil
}

// GetDocumentTags は指定されたドキュメントのタグを取得します。
func GetDocumentTags(db *sql.DB, documentID int) ([]string, error) {
	rows, err := db.Query(`
SELECT t.name
FROM tags t
JOIN document_tag_links dtl ON t.id = dtl.tag_id
WHERE dtl.document_id = ?
ORDER BY t.name ASC
`, documentID)
	if err != nil {
		return nil, fmt.Errorf("failed to query document tags: %w", err)
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tagName string
		err := rows.Scan(&tagName)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tag row: %w", err)
		}
		tags = append(tags, tagName)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return tags, nil
}

// GetDocumentFragments は指定されたドキュメントに関連するフラグメントを取得します。
func GetDocumentFragments(db *sql.DB, documentID int) ([]Fragment, error) {
	rows, err := db.Query(`
SELECT f.id, f.content, f.created_at
FROM fragments f
JOIN document_fragment_links dfl ON f.id = dfl.fragment_id
WHERE dfl.document_id = ?
ORDER BY f.created_at ASC
`, documentID)
	if err != nil {
		return nil, fmt.Errorf("failed to query document fragments: %w", err)
	}
	defer rows.Close()

	var fragments []Fragment
	for rows.Next() {
		var fragment Fragment
		err := rows.Scan(&fragment.ID, &fragment.Content, &fragment.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan fragment row: %w", err)
		}
		fragments = append(fragments, fragment)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return fragments, nil
}

// InsertDocument は新しいドキュメントをデータベースに挿入します。
func InsertDocument(db *sql.DB, title, content, summary string) (int64, error) {
	stmt, err := db.Prepare("INSERT INTO documents(title, content_markdown, summary, created_at, updated_at) VALUES(?, ?, ?, ?, ?)")
	if err != nil {
		return 0, fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	now := time.Now().Format(time.RFC3339)
	res, err := stmt.Exec(title, content, summary, now, now)
	if err != nil {
		return 0, fmt.Errorf("failed to execute statement: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert ID: %w", err)
	}
	return id, nil
}

// UpdateDocument は既存のドキュメントを更新します。
func UpdateDocument(db *sql.DB, id int, title, content, summary string) error {
	stmt, err := db.Prepare("UPDATE documents SET title = ?, content_markdown = ?, summary = ?, updated_at = ? WHERE id = ?")
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	now := time.Now().Format(time.RFC3339)
	_, err = stmt.Exec(title, content, summary, now, id)
	if err != nil {
		return fmt.Errorf("failed to execute statement: %w", err)
	}
	return nil
}

// InsertQuestion は新しい質問をデータベースに挿入します
func InsertQuestion(db *sql.DB, questionText string, contextFragmentIDs []int, contextDocumentIDs []int) (int64, error) {
	stmt, err := db.Prepare("INSERT INTO questions(question_text, context_fragment_ids, context_document_ids, created_at) VALUES(?, ?, ?, ?)")
	if err != nil {
		return 0, fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	// JSON形式に変換
	var fragmentIDsJSON, documentIDsJSON string
	if len(contextFragmentIDs) > 0 {
		fragmentIDsJSON = fmt.Sprintf("%v", contextFragmentIDs)
	}
	if len(contextDocumentIDs) > 0 {
		documentIDsJSON = fmt.Sprintf("%v", contextDocumentIDs)
	}

	res, err := stmt.Exec(questionText, fragmentIDsJSON, documentIDsJSON, time.Now().Format(time.RFC3339))
	if err != nil {
		return 0, fmt.Errorf("failed to execute statement: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert ID: %w", err)
	}
	return id, nil
}

// GetPendingQuestions は未回答の質問を取得します
func GetPendingQuestions(db *sql.DB) ([]Question, error) {
	rows, err := db.Query("SELECT id, question_text, context_fragment_ids, context_document_ids, status, created_at, answered_at, answer_fragment_id FROM questions WHERE status = 'pending' ORDER BY created_at DESC")
	if err != nil {
		return nil, fmt.Errorf("failed to query pending questions: %w", err)
	}
	defer rows.Close()

	var questions []Question
	for rows.Next() {
		var q Question
		err := rows.Scan(&q.ID, &q.QuestionText, &q.ContextFragmentIDs, &q.ContextDocumentIDs, &q.Status, &q.CreatedAt, &q.AnsweredAt, &q.AnswerFragmentID)
		if err != nil {
			return nil, fmt.Errorf("failed to scan question row: %w", err)
		}
		questions = append(questions, q)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return questions, nil
}

// GetAllQuestions は全ての質問を取得します
func GetAllQuestions(db *sql.DB) ([]Question, error) {
	rows, err := db.Query("SELECT id, question_text, context_fragment_ids, context_document_ids, status, created_at, answered_at, answer_fragment_id FROM questions ORDER BY created_at DESC")
	if err != nil {
		return nil, fmt.Errorf("failed to query questions: %w", err)
	}
	defer rows.Close()

	var questions []Question
	for rows.Next() {
		var q Question
		err := rows.Scan(&q.ID, &q.QuestionText, &q.ContextFragmentIDs, &q.ContextDocumentIDs, &q.Status, &q.CreatedAt, &q.AnsweredAt, &q.AnswerFragmentID)
		if err != nil {
			return nil, fmt.Errorf("failed to scan question row: %w", err)
		}
		questions = append(questions, q)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return questions, nil
}

// AnswerQuestion は質問に回答としてフラグメントを関連付けます
func AnswerQuestion(db *sql.DB, questionID int, answerFragmentID int64) error {
	stmt, err := db.Prepare("UPDATE questions SET status = 'answered', answered_at = ?, answer_fragment_id = ? WHERE id = ?")
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	now := time.Now().Format(time.RFC3339)
	_, err = stmt.Exec(now, answerFragmentID, questionID)
	if err != nil {
		return fmt.Errorf("failed to execute statement: %w", err)
	}
	return nil
}

// GetQuestionByID は指定されたIDの質問を取得します
func GetQuestionByID(db *sql.DB, questionID int) (*Question, error) {
	var q Question
	err := db.QueryRow("SELECT id, question_text, context_fragment_ids, context_document_ids, status, created_at, answered_at, answer_fragment_id FROM questions WHERE id = ?", questionID).
		Scan(&q.ID, &q.QuestionText, &q.ContextFragmentIDs, &q.ContextDocumentIDs, &q.Status, &q.CreatedAt, &q.AnsweredAt, &q.AnswerFragmentID)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("question not found: %d", questionID)
		}
		return nil, fmt.Errorf("failed to query question: %w", err)
	}

	return &q, nil
}

// LinkFragmentToDocument はフラグメントとドキュメントをリンクします。
func LinkFragmentToDocument(db *sql.DB, fragmentID int, documentID int64) error {
	stmt, err := db.Prepare("INSERT OR IGNORE INTO document_fragment_links(document_id, fragment_id) VALUES(?, ?)")
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(documentID, fragmentID)
	if err != nil {
		return fmt.Errorf("failed to execute statement: %w", err)
	}
	return nil
}

// InsertOrGetTag はタグを挿入するか、既存のタグのIDを取得します。
func InsertOrGetTag(db *sql.DB, tagName string) (int64, error) {
	// まず既存のタグを検索
	var tagID int64
	err := db.QueryRow("SELECT id FROM tags WHERE name = ?", tagName).Scan(&tagID)
	if err == nil {
		return tagID, nil
	}
	if err != sql.ErrNoRows {
		return 0, fmt.Errorf("failed to query existing tag: %w", err)
	}

	// 存在しない場合は新しいタグを挿入
	stmt, err := db.Prepare("INSERT INTO tags(name) VALUES(?)")
	if err != nil {
		return 0, fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(tagName)
	if err != nil {
		return 0, fmt.Errorf("failed to execute statement: %w", err)
	}

	tagID, err = res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert ID: %w", err)
	}
	return tagID, nil
}

// LinkDocumentToTag はドキュメントとタグをリンクします。
func LinkDocumentToTag(db *sql.DB, documentID int64, tagID int64) error {
	stmt, err := db.Prepare("INSERT OR IGNORE INTO document_tag_links(document_id, tag_id) VALUES(?, ?)")
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(documentID, tagID)
	if err != nil {
		return fmt.Errorf("failed to execute statement: %w", err)
	}
	return nil
}

// DeleteAllDocuments はすべてのドキュメントと関連リンクを削除します
func DeleteAllDocuments(db *sql.DB) error {
	// トランザクション開始
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// document_tag_links削除
	_, err = tx.Exec("DELETE FROM document_tag_links")
	if err != nil {
		return fmt.Errorf("failed to delete document_tag_links: %w", err)
	}

	// document_fragment_links削除
	_, err = tx.Exec("DELETE FROM document_fragment_links")
	if err != nil {
		return fmt.Errorf("failed to delete document_fragment_links: %w", err)
	}

	// documents削除
	_, err = tx.Exec("DELETE FROM documents")
	if err != nil {
		return fmt.Errorf("failed to delete documents: %w", err)
	}

	// tags削除（使われていないタグを削除）
	_, err = tx.Exec("DELETE FROM tags")
	if err != nil {
		return fmt.Errorf("failed to delete tags: %w", err)
	}

	// コミット
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteAllData はすべてのデータ（フラグメント、質問含む）を削除します
func DeleteAllData(db *sql.DB) error {
	// トランザクション開始
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// document_tag_links削除
	_, err = tx.Exec("DELETE FROM document_tag_links")
	if err != nil {
		return fmt.Errorf("failed to delete document_tag_links: %w", err)
	}

	// document_fragment_links削除
	_, err = tx.Exec("DELETE FROM document_fragment_links")
	if err != nil {
		return fmt.Errorf("failed to delete document_fragment_links: %w", err)
	}

	// documents削除
	_, err = tx.Exec("DELETE FROM documents")
	if err != nil {
		return fmt.Errorf("failed to delete documents: %w", err)
	}

	// tags削除
	_, err = tx.Exec("DELETE FROM tags")
	if err != nil {
		return fmt.Errorf("failed to delete tags: %w", err)
	}

	// questions削除
	_, err = tx.Exec("DELETE FROM questions")
	if err != nil {
		return fmt.Errorf("failed to delete questions: %w", err)
	}

	// fragments削除
	_, err = tx.Exec("DELETE FROM fragments")
	if err != nil {
		return fmt.Errorf("failed to delete fragments: %w", err)
	}

	// コミット
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
