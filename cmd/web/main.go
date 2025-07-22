package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"insight/src/ai"
	"insight/src/db"
	"insight/src/models"
	"insight/src/usecase"

	"github.com/gorilla/mux"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"gorm.io/gorm"
)

type Server struct {
	documentUsecase *usecase.DocumentUsecase
	fragmentUsecase *usecase.FragmentUsecase
	markdown        goldmark.Markdown
	templates       *template.Template
	db              *gorm.DB // データベース接続を保持
}

func main() {
	// データベース初期化
	database, err := db.Init(nil)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close(database)

	// Markdownパーサー設定
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
		),
	)

	// テンプレート読み込み
	templates, err := template.ParseGlob("web/templates/*.go.tmpl")
	if err != nil {
		log.Fatal("Failed to parse templates:", err)
	}

	// サーバー初期化
	server := &Server{
		documentUsecase: usecase.NewDocumentUsecase(database),
		fragmentUsecase: usecase.NewFragmentUsecase(database),
		markdown:        md,
		templates:       templates,
		db:              database,
	}

	// ルーター設定
	r := mux.NewRouter()
	r.HandleFunc("/", server.handleHome).Methods("GET")
	r.HandleFunc("/documents", server.handleDocuments).Methods("GET")
	r.HandleFunc("/documents/{id}", server.handleDocumentDetail).Methods("GET")
	r.HandleFunc("/fragments", server.handleFragments).Methods("GET")
	r.HandleFunc("/fragments", server.handleCreateFragment).Methods("POST")
	r.HandleFunc("/api/ai/create", server.handleAICreate).Methods("POST")
	r.HandleFunc("/api/ai/compress", server.handleAICompress).Methods("POST")
	r.HandleFunc("/api/documents/search", server.handleDocumentSearch).Methods("GET")
	r.HandleFunc("/api/documents/{id}/ask", server.handleDocumentAsk).Methods("POST")
	r.HandleFunc("/api/documents/ask", server.handleGlobalDocumentAsk).Methods("POST")

	// 静的ファイルの配信
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("web/static/"))))

	// サーバー起動
	port := "8084"
	fmt.Printf("Server starting on port %s...\n", port)
	fmt.Printf("Access: http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

// parseMarkdown はMarkdownテキストをHTMLに変換します
func (s *Server) parseMarkdown(source string) template.HTML {
	var buf bytes.Buffer
	if err := s.markdown.Convert([]byte(source), &buf); err != nil {
		return template.HTML("<p>Error parsing markdown</p>")
	}
	return template.HTML(buf.String())
}

// executeTemplateWithLogging はテンプレート実行時にログを出力します
func (s *Server) executeTemplateWithLogging(w http.ResponseWriter, templateName string, data interface{}) error {
	var buf bytes.Buffer

	// 先にバッファに出力してエラーをチェック
	if err := s.templates.ExecuteTemplate(&buf, templateName, data); err != nil {
		log.Printf("Template execution error for %s: %v", templateName, err)
		return err
	}

	// エラーがなければ実際のレスポンスに書き込み
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, writeErr := w.Write(buf.Bytes())
	return writeErr
}

func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/documents", http.StatusSeeOther)
}

func (s *Server) handleDocuments(w http.ResponseWriter, r *http.Request) {
	// 利用可能なバージョンを取得
	versions, err := s.documentUsecase.GetDistinctVersions()
	if err != nil {
		http.Error(w, "Failed to fetch versions", http.StatusInternalServerError)
		return
	}

	// バージョンパラメータの取得
	versionParam := r.URL.Query().Get("version")

	var documents []models.Document
	var selectedVersion string

	if versionParam != "" {
		// 特定のバージョンのドキュメントを取得
		// 複数の時刻フォーマットを試行
		var versionTime time.Time
		var parseErr error

		// ISO8601形式を試行
		versionTime, parseErr = time.Parse("2006-01-02T15:04:05Z", versionParam)
		if parseErr != nil {
			// RFC3339形式を試行
			versionTime, parseErr = time.Parse(time.RFC3339, versionParam)
			if parseErr != nil {
				// データベース形式を試行
				versionTime, parseErr = time.Parse("2006-01-02 15:04:05.999999-07:00", versionParam)
				if parseErr != nil {
					http.Error(w, "Invalid version format", http.StatusBadRequest)
					return
				}
			}
		}
		documents, err = s.documentUsecase.GetDocumentsByVersion(versionTime)
		selectedVersion = versionParam
	} else {
		// パラメータが指定されていない場合は最新バージョンを使用
		if len(versions) > 0 {
			latestVersion := versions[0] // versions は既に DESC でソートされている
			documents, err = s.documentUsecase.GetDocumentsByVersion(latestVersion)
			selectedVersion = latestVersion.Format("2006-01-02 15:04:05.999999-07:00")
		} else {
			// バージョンが存在しない場合は空のドキュメントリスト
			documents = []models.Document{}
		}
	}

	if err != nil {
		http.Error(w, "Failed to fetch documents", http.StatusInternalServerError)
		return
	}

	data := struct {
		Documents       []interface{}
		Versions        []time.Time
		SelectedVersion string
	}{
		Documents:       make([]interface{}, len(documents)),
		Versions:        versions,
		SelectedVersion: selectedVersion,
	}

	for i, doc := range documents {
		data.Documents[i] = doc
	}

	if err := s.executeTemplateWithLogging(w, "documents_page.go.tmpl", data); err != nil {
		http.Error(w, "Template execution error", http.StatusInternalServerError)
		return
	}
}

func (s *Server) handleDocumentDetail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid document ID", http.StatusBadRequest)
		return
	}

	document, err := s.documentUsecase.GetDocument(uint(id))
	if err != nil {
		http.Error(w, "Document not found", http.StatusNotFound)
		return
	}

	// MarkdownをHTMLに変換
	contentHTML := s.parseMarkdown(document.Content)

	data := struct {
		*models.Document
		ContentHTML template.HTML
	}{
		Document:    document,
		ContentHTML: contentHTML,
	}

	if err := s.executeTemplateWithLogging(w, "document_detail_page.go.tmpl", data); err != nil {
		http.Error(w, "Template execution error", http.StatusInternalServerError)
		return
	}
}

func (s *Server) handleFragments(w http.ResponseWriter, r *http.Request) {
	fragments, err := s.fragmentUsecase.GetAllFragments()
	if err != nil {
		http.Error(w, "Failed to fetch fragments", http.StatusInternalServerError)
		return
	}

	data := struct {
		Fragments []interface{}
	}{
		Fragments: make([]interface{}, len(fragments)),
	}

	for i, fragment := range fragments {
		data.Fragments[i] = fragment
	}

	if err := s.executeTemplateWithLogging(w, "fragments_page.go.tmpl", data); err != nil {
		http.Error(w, "Template execution error", http.StatusInternalServerError)
		return
	}
}

func (s *Server) handleCreateFragment(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	content := r.FormValue("content")
	if content == "" {
		http.Error(w, "Content is required", http.StatusBadRequest)
		return
	}

	input := usecase.CreateFragmentInput{
		Content: content,
	}

	_, err := s.fragmentUsecase.CreateFragment(input)
	if err != nil {
		http.Error(w, "Failed to create fragment", http.StatusInternalServerError)
		return
	}

	// フラグメント一覧にリダイレクト
	http.Redirect(w, r, "/fragments", http.StatusSeeOther)
}

func (s *Server) handleAICreate(w http.ResponseWriter, r *http.Request) {
	// AIサービス初期化
	aiService, err := ai.NewService(s.db)
	if err != nil {
		http.Error(w, "Failed to create AI service", http.StatusInternalServerError)
		return
	}

	// ドキュメント作成
	err = aiService.CreateDocuments(context.Background())
	if err != nil {
		http.Error(w, "Failed to create documents", http.StatusInternalServerError)
		return
	}

	// JSON レスポンス
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Documents created successfully",
	})
}

func (s *Server) handleDocumentSearch(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if query == "" {
		http.Error(w, "Query parameter 'q' is required", http.StatusBadRequest)
		return
	}

	// 全ドキュメントを取得
	documents, err := s.documentUsecase.GetAllDocuments()
	if err != nil {
		http.Error(w, "Failed to fetch documents", http.StatusInternalServerError)
		return
	}

	// クライアントサイドでの検索用に全ドキュメントをJSONで返す
	// 部分一致検索はJavaScript側で実行
	var searchableDocuments []map[string]interface{}

	for _, doc := range documents {
		searchableDocuments = append(searchableDocuments, map[string]interface{}{
			"id":                 doc.ID,
			"title":              doc.Title,
			"summary":            doc.Summary,
			"content":            doc.Content,
			"created_at":         doc.CreatedAt.Format("2006-01-02 15:04:05"),
			"version_created_at": doc.VersionCreatedAt.Format("2006-01-02 15:04:05.999999-07:00"),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"documents": searchableDocuments,
		"query":     query,
	})
}

func (s *Server) handleAICompress(w http.ResponseWriter, r *http.Request) {
	// AIサービス初期化
	aiService, err := ai.NewService(s.db)
	if err != nil {
		http.Error(w, "Failed to create AI service", http.StatusInternalServerError)
		return
	}

	// フラグメント圧縮
	err = aiService.CompressFragments(context.Background())
	if err != nil {
		http.Error(w, "Failed to compress fragments", http.StatusInternalServerError)
		return
	}

	// JSON レスポンス
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Fragments compressed successfully",
	})
}

func (s *Server) handleDocumentAsk(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid document ID", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	question := r.FormValue("question")
	if question == "" {
		http.Error(w, "Question is required", http.StatusBadRequest)
		return
	}

	useWebSearch := r.FormValue("web_search") == "true"

	// QAサービス初期化
	qaService, err := ai.NewQAService(s.db)
	if err != nil {
		http.Error(w, "Failed to create QA service", http.StatusInternalServerError)
		return
	}

	// 質問応答実行
	qaRequest := ai.QARequest{
		DocumentID:   uint(id),
		Question:     question,
		UseWebSearch: useWebSearch,
	}

	response, err := qaService.AskQuestion(context.Background(), qaRequest)
	if err != nil {
		http.Error(w, "Failed to process question", http.StatusInternalServerError)
		return
	}

	// JSON レスポンス
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) handleGlobalDocumentAsk(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	question := r.FormValue("question")
	if question == "" {
		http.Error(w, "Question is required", http.StatusBadRequest)
		return
	}

	useWebSearch := r.FormValue("web_search") == "true"

	// QAサービス初期化
	qaService, err := ai.NewQAService(s.db)
	if err != nil {
		http.Error(w, "Failed to create QA service", http.StatusInternalServerError)
		return
	}

	// 全ドキュメント質問応答実行
	qaRequest := ai.GlobalQARequest{
		Question:     question,
		UseWebSearch: useWebSearch,
	}

	response, err := qaService.AskGlobalQuestion(context.Background(), qaRequest)
	if err != nil {
		http.Error(w, "Failed to process question", http.StatusInternalServerError)
		return
	}

	// JSON レスポンス
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
