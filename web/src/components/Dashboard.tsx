'use client';

import { useState, useEffect } from 'react';
import TwitterLikeInterface from './TwitterLikeInterface';

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

export default function Dashboard() {
  const [fragments, setFragments] = useState<Fragment[]>([]);
  const [unprocessedFragments, setUnprocessedFragments] = useState<Fragment[]>([]);
  const [documents, setDocuments] = useState<Document[]>([]);
  const [allDocuments, setAllDocuments] = useState<Document[]>([]);
  const [tags, setTags] = useState<Tag[]>([]);
  
  // 検索・フィルター関連
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedTagIds, setSelectedTagIds] = useState<number[]>([]);
  
  // AI処理
  const [processLoading, setProcessLoading] = useState(false);
  const [processResult, setProcessResult] = useState<{documents: Array<{id: number, title: string, summary: string}>} | null>(null);
  
  // データ読み込み
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = async () => {
    try {
      const [fragmentsRes, unprocessedRes, documentsRes, tagsRes] = await Promise.all([
        fetch('/api/fragments'),
        fetch('/api/fragments?unprocessed=true'),
        fetch('/api/documents'),
        fetch('/api/tags'),
      ]);

      const [fragmentsData, unprocessedData, documentsData, tagsData] = await Promise.all([
        fragmentsRes.json(),
        unprocessedRes.json(),
        documentsRes.json(),
        tagsRes.json(),
      ]);

      setFragments(fragmentsData);
      setUnprocessedFragments(unprocessedData);
      setDocuments(documentsData);
      setAllDocuments(documentsData);
      setTags(tagsData);
    } catch (error) {
      console.error('Error fetching data:', error);
    } finally {
      setLoading(false);
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
        
        alert(`処理完了: ${data.documents.length}個のドキュメントが作成/更新されました。`);
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
        const searchDocuments = async () => {
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
        searchDocuments();
      } else {
        setDocuments(allDocuments);
      }
    }, 300); // デバウンス

    return () => clearTimeout(timeoutId);
  }, [searchQuery, selectedTagIds, allDocuments]);

  const unprocessedCount = unprocessedFragments.length;

  const handleFragmentCreate = async () => {
    // フラグメント作成後にデータ再取得
    await fetchData();
  };

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
    <div className="min-h-screen bg-gray-900">
      <div className="max-w-7xl mx-auto px-4 py-8">
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          
          {/* 左カラム: フラグメント入力 */}
          <div className="lg:col-span-2">
            <div className="bg-gray-800 rounded-lg shadow-sm p-6 mb-8">
              <h2 className="text-xl font-semibold mb-4 text-white">フラグメント</h2>
              <TwitterLikeInterface onFragmentCreate={handleFragmentCreate} />
            </div>
          </div>

          {/* 右カラム: 統計、AI処理、ドキュメント一覧 */}
          <div className="space-y-6">
            {/* 統計カード */}
            <div className="bg-gray-800 rounded-lg shadow-sm p-6">
              <h3 className="text-lg font-semibold mb-4 text-white">統計</h3>
              <div className="space-y-3">
                <div className="flex justify-between">
                  <span className="text-gray-300">フラグメント:</span>
                  <span className="font-medium text-white">{fragments.length}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-300">ドキュメント:</span>
                  <span className="font-medium text-white">{documents.length}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-300">タグ:</span>
                  <span className="font-medium text-white">{tags.length}</span>
                </div>
              </div>
            </div>

            {/* AI処理カード */}
            <div className="bg-gray-800 rounded-lg shadow-sm p-6">
              <h3 className="text-lg font-semibold mb-4 text-white">AI処理</h3>
              <button
                onClick={handleProcess}
                disabled={processLoading || unprocessedCount === 0}
                className="w-full px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {processLoading ? '処理中...' : `フラグメントを処理 (${unprocessedCount})`}
              </button>
              
              {processResult && (
                <div className="mt-4 p-3 bg-green-900/50 border border-green-700 rounded-md">
                  <p className="text-sm text-green-300">
                    {processResult.documents.length}個のドキュメントが作成/更新されました
                  </p>
                </div>
              )}
            </div>

            {/* ドキュメント一覧カード */}
            <div className="bg-gray-800 rounded-lg shadow-sm p-6">
              <div className="flex items-center justify-between mb-6">
                <h3 className="text-lg font-semibold text-white">ドキュメント</h3>
                <div className="flex items-center space-x-2">
                  <input
                    type="text"
                    placeholder="検索..."
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                    className="px-2 py-1 border border-gray-600 rounded text-xs bg-gray-700 text-white placeholder-gray-400 w-24"
                  />
                  {(searchQuery || selectedTagIds.length > 0) && (
                    <button
                      onClick={handleResetSearch}
                      className="text-xs text-gray-400 hover:text-gray-200"
                    >
                      リセット
                    </button>
                  )}
                </div>
              </div>

              {/* タグフィルター */}
              {tags.length > 0 && (
                <div className="mb-4">
                  <p className="text-xs text-gray-300 mb-2">タグでフィルター:</p>
                  <div className="flex flex-wrap gap-1">
                    {tags.map(tag => (
                      <button
                        key={tag.id}
                        onClick={() => handleTagToggle(tag.id)}
                        className={`px-2 py-1 rounded-full text-xs ${
                          selectedTagIds.includes(tag.id)
                            ? 'bg-blue-600 text-white'
                            : 'bg-gray-700 text-gray-300 hover:bg-gray-600'
                        }`}
                      >
                        {tag.name}
                      </button>
                    ))}
                  </div>
                </div>
              )}

              {/* ドキュメントリスト */}
              <div className="space-y-3 max-h-96 overflow-y-auto">
                {documents.length === 0 ? (
                  <p className="text-gray-400 text-center py-4 text-sm">ドキュメントがありません</p>
                ) : (
                  documents.map(doc => (
                    <div
                      key={doc.id}
                      onClick={() => handleDocumentClick(doc)}
                      className="p-3 border border-gray-700 rounded-lg hover:bg-gray-700 cursor-pointer transition-colors"
                    >
                      <h4 className="font-medium text-white mb-1 text-sm">{doc.title}</h4>
                      <p className="text-xs text-gray-300 mb-1 line-clamp-2">{doc.summary}</p>
                      <p className="text-xs text-gray-400">
                        作成: {new Date(doc.createdAt).toLocaleDateString()}
                      </p>
                    </div>
                  ))
                )}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}