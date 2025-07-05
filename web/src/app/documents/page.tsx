'use client';

import { useState, useEffect } from 'react';
import Link from 'next/link';

interface Document {
  id: number;
  title: string;
  summary: string;
  createdAt: string;
  updatedAt: string;
}

export default function DocumentsPage() {
  const [documents, setDocuments] = useState<Document[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchDocuments();
  }, []);

  const fetchDocuments = async () => {
    try {
      const response = await fetch('/api/documents');
      if (response.ok) {
        const data = await response.json();
        setDocuments(data);
      }
    } catch (error) {
      console.error('Error fetching documents:', error);
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen p-8">
        <div className="max-w-4xl mx-auto">
          <p>読み込み中...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen p-8">
      <header className="max-w-4xl mx-auto mb-8">
        <Link href="/" className="text-blue-600 hover:underline mb-4 inline-block">
          ← ホームに戻る
        </Link>
        <h1 className="text-3xl font-bold mb-2">ドキュメント一覧</h1>
        <p className="text-gray-600">生成されたドキュメントを確認できます</p>
      </header>
      
      <main className="max-w-4xl mx-auto">
        {documents.length === 0 ? (
          <div className="bg-white rounded-lg shadow-md p-6 text-center">
            <p className="text-gray-500">ドキュメントがありません</p>
            <Link href="/fragments/create" className="text-blue-600 hover:underline mt-2 inline-block">
              フラグメントを追加して始めましょう →
            </Link>
          </div>
        ) : (
          <div className="space-y-4">
            {documents.map((doc) => (
              <div key={doc.id} className="bg-white rounded-lg shadow-md p-6">
                <h2 className="text-xl font-semibold mb-2">
                  <Link href={`/documents/${doc.id}`} className="text-blue-600 hover:underline">
                    {doc.title}
                  </Link>
                </h2>
                <p className="text-gray-600 mb-2">{doc.summary}</p>
                <div className="text-sm text-gray-500">
                  作成: {new Date(doc.createdAt).toLocaleDateString('ja-JP')}
                  {doc.updatedAt !== doc.createdAt && (
                    <span className="ml-4">
                      更新: {new Date(doc.updatedAt).toLocaleDateString('ja-JP')}
                    </span>
                  )}
                </div>
              </div>
            ))}
          </div>
        )}
      </main>
    </div>
  );
}