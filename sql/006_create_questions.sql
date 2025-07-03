CREATE TABLE IF NOT EXISTS questions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    question_text TEXT NOT NULL,
    context_fragment_ids TEXT, -- JSON形式でフラグメントIDリストを保存
    context_document_ids TEXT, -- JSON形式でドキュメントIDリストを保存
    status TEXT DEFAULT 'pending', -- 'pending', 'answered', 'archived'
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    answered_at DATETIME,
    answer_fragment_id INTEGER, -- 回答として保存されたフラグメントのID
    FOREIGN KEY (answer_fragment_id) REFERENCES fragments (id)
);
