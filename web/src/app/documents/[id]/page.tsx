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
  parentId?: number;
  createdAt: string;
  updatedAt: string;
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
              <h2 className="text-lg font-semibold mb-4 text-white">関連フラグメント ({document.fragments.length}件)</h2>
              
              {/* フラグメント統計 */}
              <div className="mb-4 grid grid-cols-3 gap-4">
                <div className="bg-gray-800 rounded-lg p-3 text-center">
                  <div className="text-lg font-semibold text-green-400">
                    {document.fragments.filter(f => !f.parentId).length}
                  </div>
                  <div className="text-xs text-gray-400">元投稿</div>
                </div>
                <div className="bg-gray-800 rounded-lg p-3 text-center">
                  <div className="text-lg font-semibold text-purple-400">
                    {document.fragments.filter(f => f.parentId).length}
                  </div>
                  <div className="text-xs text-gray-400">返信</div>
                </div>
                <div className="bg-gray-800 rounded-lg p-3 text-center">
                  <div className="text-lg font-semibold text-blue-400">
                    {document.fragments.filter(f => f.url).length}
                  </div>
                  <div className="text-xs text-gray-400">URL付き</div>
                </div>
              </div>

              <div className="space-y-4">
                {document.fragments
                  .sort((a, b) => new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime())
                  .map((fragment, index) => (
                  <div key={fragment.id} className={`bg-gray-800 border rounded-lg p-4 ${fragment.parentId ? 'border-purple-600 ml-6' : 'border-gray-600'}`}>
                    {/* フラグメントヘッダー */}
                    <div className="flex items-center justify-between mb-3 pb-2 border-b border-gray-700">
                      <div className="flex items-center space-x-3">
                        <span className="inline-flex items-center justify-center w-6 h-6 rounded-full text-xs bg-gray-700 text-gray-300">
                          {index + 1}
                        </span>
                        <span className="inline-flex items-center px-2 py-1 rounded-full text-xs bg-blue-900/50 text-blue-300">
                          Fragment #{fragment.id}
                        </span>
                        {fragment.parentId && (
                          <span className="inline-flex items-center px-2 py-1 rounded-full text-xs bg-purple-900/50 text-purple-300">
                            返信 (親: #{fragment.parentId})
                          </span>
                        )}
                        {!fragment.parentId && (
                          <span className="inline-flex items-center px-2 py-1 rounded-full text-xs bg-green-900/50 text-green-300">
                            元投稿
                          </span>
                        )}
                      </div>
                      <div className="text-xs text-gray-400">
                        {new Date(fragment.createdAt).toLocaleString('ja-JP')}
                      </div>
                    </div>

                    {/* フラグメント内容 */}
                    <div className="mb-3">
                      <p className="text-gray-200 whitespace-pre-wrap leading-relaxed">{fragment.content}</p>
                    </div>

                    {/* 追加情報 */}
                    <div className="space-y-2">
                      {fragment.url && (
                        <div className="flex items-center space-x-2">
                          <svg className="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
                          </svg>
                          <a
                            href={fragment.url}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="text-blue-400 hover:text-blue-300 text-sm break-all"
                          >
                            {fragment.url}
                          </a>
                        </div>
                      )}
                      
                      {fragment.imagePath && (
                        <div className="flex items-center space-x-2">
                          <svg className="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
                          </svg>
                          <span className="text-gray-400 text-sm">画像: {fragment.imagePath}</span>
                        </div>
                      )}

                      {fragment.updatedAt !== fragment.createdAt && (
                        <div className="flex items-center space-x-2">
                          <svg className="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                          </svg>
                          <span className="text-gray-400 text-xs">
                            更新: {new Date(fragment.updatedAt).toLocaleString('ja-JP')}
                          </span>
                        </div>
                      )}
                    </div>
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