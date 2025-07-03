CREATE TABLE IF NOT EXISTS documents (
    id INTEGER PRIMARY KEY, -- 主キー
    title TEXT NOT NULL, -- AIが判断したトピック名
    content_markdown TEXT NOT NULL, -- Markdown形式の本文
    summary TEXT, -- この文書の要約 (AIが随時更新、NULL許容)
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP, -- 生成日時
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP -- 最終更新日時
);
