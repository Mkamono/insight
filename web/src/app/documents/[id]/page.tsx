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
  processed: boolean;
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
      <div className="min-h-screen p-8 bg-gray-50 dark:bg-gray-900">
        <div className="max-w-4xl mx-auto">
          <p className="text-gray-900 dark:text-white">読み込み中...</p>
        </div>
      </div>
    );
  }

  if (!document) {
    return (
      <div className="min-h-screen p-8 bg-gray-50 dark:bg-gray-900">
        <div className="max-w-4xl mx-auto">
          <p className="text-gray-900 dark:text-white">ドキュメントが見つかりません</p>
          <Link href="/" className="text-blue-600 dark:text-blue-400 hover:underline">
            ホームに戻る
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen p-8 bg-gray-50 dark:bg-gray-900">
      <header className="max-w-4xl mx-auto mb-8">
        <Link href="/" className="text-blue-600 dark:text-blue-400 hover:underline mb-4 inline-block">
          ← ホームに戻る
        </Link>
        <h1 className="text-3xl font-bold mb-2 text-gray-900 dark:text-white">{document.title}</h1>
        <div className="text-sm text-gray-500 dark:text-gray-400">
          作成: {new Date(document.createdAt).toLocaleDateString('ja-JP')}
          {document.updatedAt !== document.createdAt && (
            <span className="ml-4">
              更新: {new Date(document.updatedAt).toLocaleDateString('ja-JP')}
            </span>
          )}
        </div>
      </header>
      
      <main className="max-w-4xl mx-auto space-y-6">
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow-md p-6">
          <div className="mb-6">
            <h2 className="text-lg font-semibold mb-2 text-gray-900 dark:text-white">要約</h2>
            <p className="text-gray-700 dark:text-gray-300 bg-gray-50 dark:bg-gray-700 p-4 rounded-md">{document.summary}</p>
          </div>
          
          {/* タグ表示 */}
          {document.tags && document.tags.length > 0 && (
            <div className="mb-6">
              <h2 className="text-lg font-semibold mb-2 text-gray-900 dark:text-white">タグ</h2>
              <div className="flex flex-wrap gap-2">
                {document.tags.map((tag) => (
                  <span
                    key={tag.id}
                    className="px-3 py-1 text-sm bg-blue-100 dark:bg-blue-900/30 text-blue-800 dark:text-blue-300 rounded-full"
                  >
                    {tag.name}
                  </span>
                ))}
              </div>
            </div>
          )}
          
          <div className="mb-6">
            <h2 className="text-lg font-semibold mb-2 text-gray-900 dark:text-white">内容</h2>
            <div className="prose prose-gray dark:prose-invert max-w-none">
              <ReactMarkdown
                components={{
                  h1: ({node, ...props}) => <h1 className="text-2xl font-bold mb-4 text-gray-900 dark:text-white" {...props} />,
                  h2: ({node, ...props}) => <h2 className="text-xl font-semibold mb-3 text-gray-900 dark:text-white" {...props} />,
                  h3: ({node, ...props}) => <h3 className="text-lg font-medium mb-2 text-gray-900 dark:text-white" {...props} />,
                  p: ({node, ...props}) => <p className="mb-4 text-gray-700 dark:text-gray-300" {...props} />,
                  ul: ({node, ...props}) => <ul className="list-disc list-inside mb-4 text-gray-700 dark:text-gray-300" {...props} />,
                  ol: ({node, ...props}) => <ol className="list-decimal list-inside mb-4 text-gray-700 dark:text-gray-300" {...props} />,
                  li: ({node, ...props}) => <li className="mb-1" {...props} />,
                  strong: ({node, ...props}) => <strong className="font-bold" {...props} />,
                  em: ({node, ...props}) => <em className="italic" {...props} />,
                  code: ({node, ...props}) => <code className="bg-gray-100 dark:bg-gray-800 px-1 py-0.5 rounded text-sm" {...props} />,
                  blockquote: ({node, ...props}) => <blockquote className="border-l-4 border-gray-300 dark:border-gray-600 pl-4 italic text-gray-600 dark:text-gray-400 mb-4" {...props} />,
                  hr: ({node, ...props}) => <hr className="my-6 border-gray-200 dark:border-gray-700" {...props} />,
                  table: ({node, ...props}) => <table className="w-full border-collapse border border-gray-300 dark:border-gray-600 mb-4" {...props} />,
                  thead: ({node, ...props}) => <thead className="bg-gray-50 dark:bg-gray-800" {...props} />,
                  tbody: ({node, ...props}) => <tbody {...props} />,
                  tr: ({node, ...props}) => <tr className="border-b border-gray-200 dark:border-gray-700" {...props} />,
                  th: ({node, ...props}) => <th className="border border-gray-300 dark:border-gray-600 px-3 py-2 text-left font-semibold text-gray-900 dark:text-white" {...props} />,
                  td: ({node, ...props}) => <td className="border border-gray-300 dark:border-gray-600 px-3 py-2 text-gray-700 dark:text-gray-300" {...props} />,
                }}
              >
                {document.content}
              </ReactMarkdown>
            </div>
          </div>
          
          {/* 参考フラグメント表示 */}
          {document.fragments && document.fragments.length > 0 && (
            <div>
              <h2 className="text-lg font-semibold mb-2 text-gray-900 dark:text-white">参考フラグメント</h2>
              <div className="space-y-3">
                {document.fragments.map((fragment) => (
                  <div
                    key={fragment.id}
                    className="p-3 bg-gray-50 dark:bg-gray-700 rounded-md border-l-4 border-gray-300 dark:border-gray-600"
                  >
                    <p className="text-gray-700 dark:text-gray-300 text-sm">{fragment.content}</p>
                    {fragment.url && (
                      <a
                        href={fragment.url}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="text-blue-600 dark:text-blue-400 hover:underline mt-2 inline-block text-sm"
                      >
                        🔗 {fragment.url}
                      </a>
                    )}
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      </main>
    </div>
  );
}