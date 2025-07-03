CREATE TABLE IF NOT EXISTS document_fragment_links (
    document_id INTEGER NOT NULL, -- documentsテーブルへの外部キー
    fragment_id INTEGER NOT NULL, -- fragmentsテーブルへの外部キー
    PRIMARY KEY (document_id, fragment_id), -- 複合主キー
    FOREIGN KEY (document_id) REFERENCES documents(id),
    FOREIGN KEY (fragment_id) REFERENCES fragments(id)
);
