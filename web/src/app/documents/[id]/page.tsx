'use client';

import { useState, useEffect } from 'react';
import Link from 'next/link';
import { useParams } from 'next/navigation';
import ReactMarkdown from 'react-markdown';

interface Tag {
  id: number;
  name: string;
  createdAt: string;
  updatedAt: string;
}

interface Fragment {
  id: number;
  content: string;
  url?: string;
  imagePath?: string;
  createdAt: string;
}

interface Document {
  id: number;
  title: string;
  content: string;
  summary: string;
  createdAt: string;
  updatedAt: string;
  tags?: Tag[];
  fragments?: Fragment[];
}

export default function DocumentDetailPage() {
  const params = useParams();
  const [document, setDocument] = useState<Document | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (params.id) {
      fetchDocument(params.id as string);
    }
  }, [params.id]);

  const fetchDocument = async (id: string) => {
    try {
      const response = await fetch(`/api/documents/${id}`);
      if (response.ok) {
        const data = await response.json();
        setDocument(data);
      }
    } catch (error) {
      console.error('Error fetching document:', error);
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen p-8 bg-gray-900">
        <div className="max-w-4xl mx-auto">
          <p className="text-white">読み込み中...</p>
        </div>
      </div>
    );
  }

  if (!document) {
    return (
      <div className="min-h-screen p-8 bg-gray-900">
        <div className="max-w-4xl mx-auto">
          <h1 className="text-2xl font-bold mb-4 text-white">ドキュメントが見つかりません</h1>
          <Link href="/" className="text-blue-400 hover:underline">
            ホームに戻る
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen p-8 bg-gray-900">
      <div className="max-w-4xl mx-auto">
        <nav className="mb-6">
          <Link href="/" className="text-blue-400 hover:underline">
            ← ダッシュボードに戻る
          </Link>
        </nav>
        
        <h1 className="text-3xl font-bold mb-2 text-white">{document.title}</h1>
        <p className="text-gray-400 text-sm mb-6">
          作成: {new Date(document.createdAt).toLocaleString()}
        </p>

        <div className="space-y-6">
          {/* 要約 */}
          {document.summary && (
            <div className="bg-gray-800 rounded-lg p-6">
              <h2 className="text-lg font-semibold mb-2 text-white">要約</h2>
              <p className="text-gray-300 bg-gray-700 p-4 rounded-md">{document.summary}</p>
            </div>
          )}

          {/* タグ */}
          {document.tags && document.tags.length > 0 && (
            <div>
              <h2 className="text-lg font-semibold mb-2 text-white">タグ</h2>
              <div className="flex flex-wrap gap-2">
                {document.tags.map(tag => (
                  <span 
                    key={tag.id} 
                    className="px-3 py-1 rounded-full text-sm bg-blue-900/50 text-blue-300"
                  >
                    {tag.name}
                  </span>
                ))}
              </div>
            </div>
          )}
          
          <div className="mb-6">
            <h2 className="text-lg font-semibold mb-2 text-white">内容</h2>
            <div className="prose prose-invert max-w-none">
              <ReactMarkdown
                components={{
                  h1: (props) => <h1 className="text-2xl font-bold mb-4 text-white" {...props} />,
                  h2: (props) => <h2 className="text-xl font-semibold mb-3 text-white" {...props} />,
                  h3: (props) => <h3 className="text-lg font-medium mb-2 text-white" {...props} />,
                  p: (props) => <p className="mb-4 text-gray-300" {...props} />,
                  ul: (props) => <ul className="list-disc list-inside mb-4 text-gray-300" {...props} />,
                  ol: (props) => <ol className="list-decimal list-inside mb-4 text-gray-300" {...props} />,
                  li: (props) => <li className="mb-1" {...props} />,
                  strong: (props) => <strong className="font-bold" {...props} />,
                  em: (props) => <em className="italic" {...props} />,
                  code: (props) => <code className="bg-gray-800 px-1 py-0.5 rounded text-sm" {...props} />,
                  blockquote: (props) => <blockquote className="border-l-4 border-gray-600 pl-4 italic text-gray-400 mb-4" {...props} />,
                  hr: (props) => <hr className="my-6 border-gray-700" {...props} />,
                  table: (props) => <table className="w-full border-collapse border border-gray-600 mb-4" {...props} />,
                  thead: (props) => <thead className="bg-gray-800" {...props} />,
                  tbody: (props) => <tbody {...props} />,
                  tr: (props) => <tr className="border-b border-gray-700" {...props} />,
                  th: (props) => <th className="border border-gray-600 px-3 py-2 text-left font-semibold text-white" {...props} />,
                  td: (props) => <td className="border border-gray-600 px-3 py-2 text-gray-300" {...props} />,
                }}
              >
                {document.content}
              </ReactMarkdown>
            </div>
          </div>

          {/* フラグメント */}
          {document.fragments && document.fragments.length > 0 && (
            <div>
              <h2 className="text-lg font-semibold mb-2 text-white">関連フラグメント</h2>
              <div className="space-y-3">
                {document.fragments.map(fragment => (
                  <div key={fragment.id} className="p-3 bg-gray-700 border-gray-600 rounded-md">
                    <p className="text-gray-300 text-sm">{fragment.content}</p>
                    {fragment.url && (
                      <div className="mt-2">
                        <a
                          href={fragment.url}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="text-blue-400 hover:text-blue-300 text-xs"
                        >
                          {fragment.url}
                        </a>
                      </div>
                    )}
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}