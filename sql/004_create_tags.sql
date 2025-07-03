CREATE TABLE IF NOT EXISTS tags (
    id INTEGER PRIMARY KEY, -- 主キー
    name TEXT NOT NULL UNIQUE -- タグ名 (例: "Python", "アイデア")
);
