package web

import (
	"database/sql"
	"insight/internal/database"
	"log"
	"net/http"
)

type Server struct {
	db *sql.DB
}

func NewServer(db *sql.DB) *Server {
	return &Server{db: db}
}

func (s *Server) Start(port string) {
	// 簡単なWebサーバーを実装
	http.HandleFunc("/", s.handleHome)
	http.HandleFunc("/documents", s.handleDocuments)

	log.Printf("Starting web server on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Insight Knowledge Base - Web interface not fully implemented"))
}

func (s *Server) handleDocuments(w http.ResponseWriter, r *http.Request) {
	documents, err := database.GetAllDocuments(s.db)
	if err != nil {
		http.Error(w, "Failed to get documents", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	for _, doc := range documents {
		w.Write([]byte(doc.Title + "\n"))
	}
}
