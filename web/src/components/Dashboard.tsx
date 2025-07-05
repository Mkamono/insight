'use client';

import { useState, useEffect } from 'react';

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
  summary: string;
  createdAt: string;
  updatedAt: string;
}

interface Tag {
  id: number;
  name: string;
  createdAt: string;
  updatedAt: string;
}

interface DocumentDetail extends Document {
  content: string;
  tags?: Tag[];
  fragments?: Fragment[];
}

export default function Dashboard() {
  const [fragments, setFragments] = useState<Fragment[]>([]);
  const [documents, setDocuments] = useState<Document[]>([]);
  const [allDocuments, setAllDocuments] = useState<Document[]>([]);
  const [tags, setTags] = useState<Tag[]>([]);
  
  // 検索・フィルター関連
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedTagIds, setSelectedTagIds] = useState<number[]>([]);
  const [showSearch, setShowSearch] = useState(false);
  
  // フラグメント追加フォーム
  const [content, setContent] = useState('');
  const [url, setUrl] = useState('');
  const [imagePath, setImagePath] = useState('');
  const [createLoading, setCreateLoading] = useState(false);
  
  // AI処理
  const [processLoading, setProcessLoading] = useState(false);
  const [processResult, setProcessResult] = useState<any>(null);
  
  // データ読み込み
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = async () => {
    try {
      const [fragmentsRes, documentsRes, tagsRes] = await Promise.all([
        fetch('/api/fragments'),
        fetch('/api/documents'),
        fetch('/api/tags')
      ]);
      
      if (fragmentsRes.ok && documentsRes.ok && tagsRes.ok) {
        const fragmentsData = await fragmentsRes.json();
        const documentsData = await documentsRes.json();
        const tagsData = await tagsRes.json();
        setFragments(fragmentsData);
        setDocuments(documentsData);
        setAllDocuments(documentsData);
        setTags(tagsData);
      }
    } catch (error) {
      console.error('Error fetching data:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleCreateFragment = async (e: React.FormEvent) => {
    e.preventDefault();
    setCreateLoading(true);

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
        setContent('');
        setUrl('');
        setImagePath('');
        await fetchData(); // データを再取得
      } else {
        alert('エラーが発生しました');
      }
    } catch (error) {
      console.error('Error:', error);
      alert('エラーが発生しました');
    } finally {
      setCreateLoading(false);
    }
  };

  const handleProcess = async () => {
    setProcessLoading(true);
    setProcessResult(null);

    try {
      const response = await fetch('/api/process', {
        method: 'POST',
      });

      if (response.ok) {
        const data = await response.json();
        setProcessResult(data);
        
        // データを再取得して最新状態に更新
        await fetchData();
        
        // フラグメントデータを再取得して確認
        const fragmentsResponse = await fetch('/api/fragments');
        if (fragmentsResponse.ok) {
          const updatedFragments = await fragmentsResponse.json();
          console.log('Processing completed. Updated fragments:', updatedFragments.length);
          console.log('Unprocessed fragments remaining:', 
            updatedFragments.filter((f: Fragment) => !f.processed).length);
          
          // 処理に失敗したフラグメントがあれば表示
          const stillUnprocessed = updatedFragments.filter((f: Fragment) => !f.processed);
          if (stillUnprocessed.length > 0) {
            console.warn('Still unprocessed fragments:', stillUnprocessed.map((f: Fragment) => ({ id: f.id, content: f.content.slice(0, 50) })));
            
            // レート制限の可能性をユーザーに通知
            if (data.documents.length < unprocessedCount) {
              alert(`処理完了: ${data.documents.length}個のドキュメントが作成/更新されました。\n\n残り${stillUnprocessed.length}個のフラグメントが未処理です。\nこれはAPIのレート制限が原因の可能性があります。\n\n少し時間をおいてから再度実行してください。`);
            }
          }
        }
      } else {
        const errorData = await response.json();
        console.error('Process error:', errorData);
        alert(`エラーが発生しました: ${errorData.error || '不明なエラー'}`);
      }
    } catch (error) {
      console.error('Error:', error);
      alert('エラーが発生しました');
    } finally {
      setProcessLoading(false);
    }
  };

  const handleDocumentClick = (doc: Document) => {
    // 別ページで開く
    window.open(`/documents/${doc.id}`, '_blank');
  };

  const handleSearch = async () => {
    try {
      const params = new URLSearchParams();
      if (searchQuery.trim()) {
        params.append('q', searchQuery.trim());
      }
      if (selectedTagIds.length > 0) {
        params.append('tags', selectedTagIds.join(','));
      }

      const response = await fetch(`/api/documents/search?${params.toString()}`);
      if (response.ok) {
        const searchResults = await response.json();
        setDocuments(searchResults);
      }
    } catch (error) {
      console.error('Error searching documents:', error);
    }
  };

  const handleResetSearch = () => {
    setSearchQuery('');
    setSelectedTagIds([]);
    setDocuments(allDocuments);
  };

  const handleTagToggle = (tagId: number) => {
    setSelectedTagIds(prev => 
      prev.includes(tagId) 
        ? prev.filter(id => id !== tagId)
        : [...prev, tagId]
    );
  };

  // 検索条件が変わったら自動検索
  useEffect(() => {
    const timeoutId = setTimeout(() => {
      if (searchQuery.trim() || selectedTagIds.length > 0) {
        handleSearch();
      } else {
        setDocuments(allDocuments);
      }
    }, 300); // デバウンス

    return () => clearTimeout(timeoutId);
  }, [searchQuery, selectedTagIds, allDocuments]);

  const unprocessedCount = fragments.filter(f => !f.processed).length;

  if (loading) {
    return (
      <div className="min-h-screen p-8">
        <div className="max-w-6xl mx-auto">
          <p>読み込み中...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen p-8 bg-gray-50 dark:bg-gray-900">
      <header className="max-w-6xl mx-auto mb-8">
        <h1 className="text-3xl font-bold mb-2 text-gray-900 dark:text-white">Insight</h1>
        <p className="text-gray-600 dark:text-gray-400">AI知識ドキュメント化システム</p>
      </header>
      
      <main className="max-w-6xl mx-auto">
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
          
          {/* 左側: フラグメント追加 + AI処理 */}
          <div className="space-y-6">
            
            {/* フラグメント追加セクション */}
            <div className="bg-white dark:bg-gray-800 rounded-lg shadow-md p-6">
              <h2 className="text-xl font-semibold mb-4 text-gray-900 dark:text-white">フラグメント追加</h2>
              <form onSubmit={handleCreateFragment} className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                    内容 *
                  </label>
                  <textarea
                    value={content}
                    onChange={(e) => setContent(e.target.value)}
                    required
                    className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-white rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                    rows={3}
                    placeholder="フラグメントの内容を入力..."
                  />
                </div>
                
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                      URL (オプション)
                    </label>
                    <input
                      type="url"
                      value={url}
                      onChange={(e) => setUrl(e.target.value)}
                      className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-white rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                      placeholder="https://..."
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                      画像パス (オプション)
                    </label>
                    <input
                      type="text"
                      value={imagePath}
                      onChange={(e) => setImagePath(e.target.value)}
                      className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-white rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                      placeholder="/path/to/image.jpg"
                    />
                  </div>
                </div>
                
                <button
                  type="submit"
                  disabled={createLoading || !content.trim()}
                  className="w-full bg-blue-600 text-white py-2 px-4 rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50"
                >
                  {createLoading ? '作成中...' : 'フラグメントを作成'}
                </button>
              </form>
            </div>

            {/* AI処理セクション */}
            <div className="bg-white dark:bg-gray-800 rounded-lg shadow-md p-6">
              <h2 className="text-xl font-semibold mb-4 text-gray-900 dark:text-white">AI処理</h2>
              <div className="mb-4">
                <p className="text-gray-600 dark:text-gray-400">
                  未処理のフラグメント: <span className="font-medium">{unprocessedCount}個</span>
                </p>
              </div>

              {/* 未処理フラグメント一覧 */}
              {unprocessedCount > 0 && (
                <div className="mb-4">
                  <h3 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">未処理フラグメント:</h3>
                  <div className="space-y-2 max-h-48 overflow-y-auto">
                    {fragments.filter(f => !f.processed).map((fragment) => (
                      <div
                        key={fragment.id}
                        className="p-2 bg-yellow-50 dark:bg-yellow-900/20 rounded-md border-l-2 border-yellow-400 dark:border-yellow-500"
                      >
                        <p className="text-xs text-gray-700 dark:text-gray-300">
                          {fragment.content.length > 100 
                            ? `${fragment.content.slice(0, 100)}...` 
                            : fragment.content}
                        </p>
                        {fragment.url && (
                          <a
                            href={fragment.url}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="text-blue-600 dark:text-blue-400 hover:underline mt-1 inline-block text-xs"
                          >
                            🔗 {fragment.url}
                          </a>
                        )}
                        <div className="text-xs text-gray-500 dark:text-gray-400 mt-1">
                          ID: {fragment.id} • {new Date(fragment.createdAt).toLocaleDateString('ja-JP')}
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              )}
              
              <button
                onClick={handleProcess}
                disabled={processLoading || unprocessedCount === 0}
                className="w-full bg-green-600 text-white py-2 px-4 rounded-md hover:bg-green-700 focus:outline-none focus:ring-2 focus:ring-green-500 disabled:opacity-50"
              >
                {processLoading ? '処理中...' : 'AI処理を実行'}
              </button>
              
              {processResult && (
                <div className="mt-4 p-4 bg-green-50 dark:bg-green-900/20 rounded-md">
                  <h3 className="font-semibold text-green-800 dark:text-green-400">処理完了</h3>
                  <p className="text-green-700 dark:text-green-300">
                    {processResult.documents.length}個のドキュメントが作成/更新されました
                  </p>
                </div>
              )}
            </div>
          </div>

          {/* 右側: ドキュメント一覧 + 詳細 */}
          <div className="space-y-6">
            
            {/* ドキュメント一覧セクション */}
            <div className="bg-white dark:bg-gray-800 rounded-lg shadow-md p-6">
              <div className="flex justify-between items-center mb-4">
                <h2 className="text-xl font-semibold text-gray-900 dark:text-white">ドキュメント一覧</h2>
                <button
                  onClick={() => setShowSearch(!showSearch)}
                  className="px-3 py-1 text-sm bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors"
                >
                  {showSearch ? '検索を隠す' : '検索・フィルター'}
                </button>
              </div>

              {/* 検索・フィルター UI */}
              {showSearch && (
                <div className="mb-4 p-4 bg-gray-50 dark:bg-gray-700 rounded-md space-y-3">
                  {/* 検索ボックス */}
                  <div>
                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                      検索
                    </label>
                    <input
                      type="text"
                      value={searchQuery}
                      onChange={(e) => setSearchQuery(e.target.value)}
                      placeholder="タイトル、内容、要約を検索..."
                      className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 dark:bg-gray-800 dark:text-white rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm"
                    />
                  </div>

                  {/* タグフィルター */}
                  {tags.length > 0 && (
                    <div>
                      <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                        タグフィルター
                      </label>
                      <div className="flex flex-wrap gap-2">
                        {tags.map((tag) => (
                          <button
                            key={tag.id}
                            onClick={() => handleTagToggle(tag.id)}
                            className={`px-2 py-1 text-xs rounded-full transition-colors ${
                              selectedTagIds.includes(tag.id)
                                ? 'bg-blue-600 text-white'
                                : 'bg-gray-200 dark:bg-gray-600 text-gray-700 dark:text-gray-300 hover:bg-gray-300 dark:hover:bg-gray-500'
                            }`}
                          >
                            {tag.name}
                          </button>
                        ))}
                      </div>
                    </div>
                  )}

                  {/* リセットボタン */}
                  {(searchQuery.trim() || selectedTagIds.length > 0) && (
                    <div className="flex justify-between items-center text-sm">
                      <span className="text-gray-600 dark:text-gray-400">
                        {documents.length}件のドキュメントが見つかりました
                      </span>
                      <button
                        onClick={handleResetSearch}
                        className="text-blue-600 dark:text-blue-400 hover:underline"
                      >
                        検索をリセット
                      </button>
                    </div>
                  )}
                </div>
              )}
              
              {documents.length === 0 ? (
                <p className="text-gray-500 dark:text-gray-400 text-center py-4">
                  ドキュメントがありません
                </p>
              ) : (
                <div className="space-y-2 max-h-96 overflow-y-auto">
                  {documents.map((doc) => (
                    <div
                      key={doc.id}
                      onClick={() => handleDocumentClick(doc)}
                      className="p-3 rounded-md cursor-pointer transition-colors bg-gray-50 dark:bg-gray-700 hover:bg-gray-100 dark:hover:bg-gray-600"
                    >
                      <h3 className="font-medium text-sm text-gray-900 dark:text-white">{doc.title}</h3>
                      <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
                        {new Date(doc.updatedAt).toLocaleDateString('ja-JP')} • 新しいタブで開く
                      </p>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>
        </div>
      </main>
    </div>
  );
}