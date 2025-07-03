CREATE TABLE IF NOT EXISTS document_tag_links (
    document_id INTEGER NOT NULL, -- documentsテーブルへの外部キー
    tag_id INTEGER NOT NULL, -- tagsテーブルへの外部キー
    PRIMARY KEY (document_id, tag_id), -- 複合主キー
    FOREIGN KEY (document_id) REFERENCES documents(id),
    FOREIGN KEY (tag_id) REFERENCES tags(id)
);
