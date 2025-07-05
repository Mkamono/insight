'use client';

import dynamic from 'next/dynamic';

const Dashboard = dynamic(() => import('../components/Dashboard'), {
  ssr: false,
  loading: () => (
    <div className="min-h-screen p-8 bg-gray-50">
      <div className="max-w-6xl mx-auto">
        <header className="mb-8">
          <h1 className="text-3xl font-bold mb-2">Insight</h1>
          <p className="text-gray-600">AI知識ドキュメント化システム</p>
        </header>
        <div className="flex items-center justify-center">
          <p>読み込み中...</p>
        </div>
      </div>
    </div>
  ),
});

export default function Home() {
  return <Dashboard />;
}
