CREATE TABLE IF NOT EXISTS fragments (
    id INTEGER PRIMARY KEY, -- 主キー
    content TEXT NOT NULL, -- つぶやきの内容 (JSON形式)
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP -- 作成日時
);
