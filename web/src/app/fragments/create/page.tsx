'use client';

import { useState } from 'react';
import Link from 'next/link';

export default function CreateFragment() {
  const [content, setContent] = useState('');
  const [url, setUrl] = useState('');
  const [imagePath, setImagePath] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);

    try {
      const response = await fetch('/api/fragments', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          content,
          url: url || null,
          imagePath: imagePath || null,
        }),
      });

      if (response.ok) {
        alert('フラグメントが作成されました！');
        setContent('');
        setUrl('');
        setImagePath('');
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
        <h1 className="text-3xl font-bold mb-2">フラグメント追加</h1>
        <p className="text-gray-600">新しい知識フラグメントを追加します</p>
      </header>
      
      <main className="max-w-4xl mx-auto">
        <form onSubmit={handleSubmit} className="bg-white rounded-lg shadow-md p-6">
          <div className="mb-6">
            <label htmlFor="content" className="block text-sm font-medium text-gray-700 mb-2">
              内容 *
            </label>
            <textarea
              id="content"
              value={content}
              onChange={(e) => setContent(e.target.value)}
              required
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              rows={6}
              placeholder="フラグメントの内容を入力してください..."
            />
          </div>
          
          <div className="mb-6">
            <label htmlFor="url" className="block text-sm font-medium text-gray-700 mb-2">
              URL（オプション）
            </label>
            <input
              type="url"
              id="url"
              value={url}
              onChange={(e) => setUrl(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              placeholder="https://example.com"
            />
          </div>
          
          <div className="mb-6">
            <label htmlFor="imagePath" className="block text-sm font-medium text-gray-700 mb-2">
              画像パス（オプション）
            </label>
            <input
              type="text"
              id="imagePath"
              value={imagePath}
              onChange={(e) => setImagePath(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              placeholder="/path/to/image.jpg"
            />
          </div>
          
          <button
            type="submit"
            disabled={loading || !content.trim()}
            className="w-full bg-blue-600 text-white py-2 px-4 rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {loading ? '作成中...' : 'フラグメントを作成'}
          </button>
        </form>
      </main>
    </div>
  );
}