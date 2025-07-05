'use client';

import { useState } from 'react';
import Link from 'next/link';

export default function ProcessPage() {
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<any>(null);

  const handleProcess = async () => {
    setLoading(true);
    setResult(null);

    try {
      const response = await fetch('/api/process', {
        method: 'POST',
      });

      if (response.ok) {
        const data = await response.json();
        setResult(data);
      } else {
        alert('エラーが発生しました');
      }
    } catch (error) {
      console.error('Error:', error);
      alert('エラーが発生しました');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen p-8">
      <header className="max-w-4xl mx-auto mb-8">
        <Link href="/" className="text-blue-600 hover:underline mb-4 inline-block">
          ← ホームに戻る
        </Link>
        <h1 className="text-3xl font-bold mb-2">AI処理</h1>
        <p className="text-gray-600">未処理のフラグメントからドキュメントを生成します</p>
      </header>
      
      <main className="max-w-4xl mx-auto">
        <div className="bg-white rounded-lg shadow-md p-6">
          <div className="mb-6">
            <button
              onClick={handleProcess}
              disabled={loading}
              className="bg-green-600 text-white py-3 px-6 rounded-md hover:bg-green-700 focus:outline-none focus:ring-2 focus:ring-green-500 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {loading ? '処理中...' : '未処理フラグメントを処理'}
            </button>
          </div>
          
          {result && (
            <div className="mt-6 p-4 bg-gray-50 rounded-md">
              <h3 className="text-lg font-semibold mb-2">処理結果</h3>
              <p className="mb-2">
                {result.documents.length}個のドキュメントが作成/更新されました
              </p>
              
              {result.documents.length > 0 && (
                <div>
                  <h4 className="font-medium mb-2">作成/更新されたドキュメント:</h4>
                  <ul className="space-y-1">
                    {result.documents.map((doc: any, index: number) => (
                      <li key={index} className="text-sm">
                        • {doc.title}
                      </li>
                    ))}
                  </ul>
                </div>
              )}
              
              <div className="mt-4">
                <Link href="/documents" className="text-blue-600 hover:underline">
                  ドキュメント一覧を見る →
                </Link>
              </div>
            </div>
          )}
        </div>
      </main>
    </div>
  );
}