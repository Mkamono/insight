'use client';

import React, { useEffect, useState } from 'react';

interface Fragment {
  id: number;
  content: string;
  url?: string;
  imagePath?: string;
  parentId?: number;
  createdAt?: string;
  updatedAt?: string;
  children?: Fragment[];
}

interface TwitterLikeInterfaceProps {
  onFragmentCreate?: (fragment: Fragment) => void;
}

export default function TwitterLikeInterface({ onFragmentCreate }: TwitterLikeInterfaceProps) {
  const [fragments, setFragments] = useState<Fragment[]>([]);
  const [loading, setLoading] = useState(false);
  const [replyTo, setReplyTo] = useState<number | null>(null);
  const [currentPage, setCurrentPage] = useState(1);
  const [itemsPerPage] = useState(10);

  useEffect(() => {
    fetchRootFragments();
  }, []);

  const fetchRootFragments = async () => {
    try {
      // 全フラグメントを取得
      const allResponse = await fetch('/api/fragments');
      const allFragments = await allResponse.json();

      // 親子関係を再帰的に構築する関数
      const buildChildren = (parentId: number): Fragment[] => {
        return allFragments
          .filter((f: Fragment) => f.parentId === parentId)
          .sort((a: Fragment, b: Fragment) => new Date(a.createdAt || '').getTime() - new Date(b.createdAt || '').getTime()) // 子は古い順（会話の流れ）
          .map((child: Fragment) => ({
            ...child,
            children: buildChildren(child.id)
          }));
      };

      // ルートフラグメント（parentId がない）に子を追加
      const rootFragments = allFragments
        .filter((f: Fragment) => !f.parentId)
        .sort((a: Fragment, b: Fragment) => new Date(b.createdAt || '').getTime() - new Date(a.createdAt || '').getTime()) // 最新順
        .map((fragment: Fragment) => ({
          ...fragment,
          children: buildChildren(fragment.id)
        }));

      setFragments(rootFragments);
    } catch (error) {
      console.error('Error fetching fragments:', error);
    }
  };


  const createFragment = async (content: string, parentId?: number) => {
    if (!content.trim()) return;

    setLoading(true);
    try {
      const response = await fetch('/api/fragments', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          content: content.trim(),
          parentId: parentId || null,
        }),
      });

      if (response.ok) {
        const newFragment = await response.json();
        onFragmentCreate?.(newFragment);

        // フラグメント作成後は全体を再読み込み
        await fetchRootFragments();

        setReplyTo(null);
        setCurrentPage(1); // 新しいフラグメントは最初のページに表示
      }
    } catch (error) {
      console.error('Error creating fragment:', error);
    } finally {
      setLoading(false);
    }
  };


  // ページネーション計算
  const totalPages = Math.ceil(fragments.length / itemsPerPage);
  const startIndex = (currentPage - 1) * itemsPerPage;
  const endIndex = startIndex + itemsPerPage;
  const currentFragments = fragments.slice(startIndex, endIndex);

  const handlePageChange = (page: number) => {
    setCurrentPage(page);
    setReplyTo(null); // ページ変更時は返信をクリア
  };

  return (
    <div className="max-w-2xl mx-auto">
      {/* メイン投稿フォーム */}
      <FragmentForm
        onSubmit={(content) => createFragment(content)}
        loading={loading}
        placeholder="何について考えていますか？"
      />

      {/* フラグメント統計 */}
      {fragments.length > 0 && (
        <div className="mt-4 text-sm text-gray-400 text-center">
          全{fragments.length}件のフラグメント
          {totalPages > 1 && (
            <span className="ml-2">
              (ページ {currentPage} / {totalPages})
            </span>
          )}
        </div>
      )}

      {/* フラグメント一覧 */}
      <div className="mt-6 space-y-4">
        {currentFragments.map(fragment => (
          <FragmentCard
            key={fragment.id}
            fragment={fragment}
            onReply={(content) => createFragment(content, fragment.id)}
            replyTo={replyTo}
            setReplyTo={setReplyTo}
            loading={loading}
          />
        ))}
      </div>

      {/* ページネーション */}
      {totalPages > 1 && (
        <div className="mt-6 flex justify-center items-center space-x-2">
          <button
            onClick={() => handlePageChange(currentPage - 1)}
            disabled={currentPage === 1}
            className="px-3 py-1 bg-gray-700 text-white rounded hover:bg-gray-600 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            前
          </button>
          
          {Array.from({ length: totalPages }, (_, i) => i + 1).map(page => (
            <button
              key={page}
              onClick={() => handlePageChange(page)}
              className={`px-3 py-1 rounded ${
                page === currentPage
                  ? 'bg-blue-600 text-white'
                  : 'bg-gray-700 text-gray-300 hover:bg-gray-600'
              }`}
            >
              {page}
            </button>
          ))}
          
          <button
            onClick={() => handlePageChange(currentPage + 1)}
            disabled={currentPage === totalPages}
            className="px-3 py-1 bg-gray-700 text-white rounded hover:bg-gray-600 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            次
          </button>
        </div>
      )}
    </div>
  );
}

interface FragmentFormProps {
  onSubmit: (content: string) => void;
  loading: boolean;
  placeholder?: string;
}

function FragmentForm({ onSubmit, loading, placeholder = "返信を書く..." }: FragmentFormProps) {
  const [content, setContent] = useState('');

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSubmit(content);
    setContent('');
  };

  return (
    <form onSubmit={handleSubmit} className="bg-gray-700 rounded-lg border border-gray-600 p-4">
      <textarea
        value={content}
        onChange={(e) => setContent(e.target.value)}
        placeholder={placeholder}
        className="w-full p-3 border border-gray-500 rounded-lg resize-none focus:outline-none focus:ring-2 focus:ring-blue-500 bg-gray-800 text-white placeholder-gray-400"
        rows={3}
        disabled={loading}
      />
      <div className="flex justify-end mt-3">
        <button
          type="submit"
          disabled={!content.trim() || loading}
          className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {loading ? '投稿中...' : '投稿'}
        </button>
      </div>
    </form>
  );
}

interface FragmentCardProps {
  fragment: Fragment;
  onReply: (content: string) => void;
  replyTo: number | null;
  setReplyTo: (id: number | null) => void;
  loading: boolean;
}

function FragmentCard({ fragment, onReply, replyTo, setReplyTo, loading }: FragmentCardProps) {
  const [showChildren, setShowChildren] = useState(true); // 常に展開

  const handleReplyClick = () => {
    setReplyTo(replyTo === fragment.id ? null : fragment.id);
  };

  const handleToggleChildren = () => {
    setShowChildren(!showChildren);
  };

  // このフラグメントまたはその子孫が返信対象かどうかチェック
  const isReplyTargetInThisTree = (targetId: number | null, currentFragment: Fragment): boolean => {
    if (!targetId) return false;
    if (currentFragment.id === targetId) return true;
    if (currentFragment.children) {
      return currentFragment.children.some(child => isReplyTargetInThisTree(targetId, child));
    }
    return false;
  };

  return (
    <div className="bg-gray-700 rounded-lg border border-gray-600 p-4">
      <div className="mb-2">
        <div className="flex items-center justify-between mb-2">
          <span className="text-xs text-gray-400">
            Fragment #{fragment.id}
          </span>
          <span className="text-xs text-gray-400">
            {fragment.createdAt && new Date(fragment.createdAt).toLocaleString()}
          </span>
        </div>

        <p className="text-gray-100 mb-3 whitespace-pre-wrap">{fragment.content}</p>

        {fragment.url && (
          <div className="mb-3">
            <a
              href={fragment.url}
              target="_blank"
              rel="noopener noreferrer"
              className="text-blue-400 hover:text-blue-300 text-sm"
            >
              {fragment.url}
            </a>
          </div>
        )}

        <div className="flex items-center space-x-4 text-gray-400">
          <button
            onClick={handleReplyClick}
            className="flex items-center space-x-1 hover:text-blue-400 text-sm"
          >
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 10h10a8 8 0 018 8v2M3 10l6 6m-6-6l6-6" />
            </svg>
            <span>返信</span>
          </button>

          {fragment.children && fragment.children.length > 0 && (
            <button
              onClick={handleToggleChildren}
              className="flex items-center space-x-1 hover:text-green-400 text-sm"
            >
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 13l-7 7-7-7m14-8l-7 7-7-7" />
              </svg>
              <span>{showChildren ? '折りたたむ' : '展開'}</span>
            </button>
          )}
        </div>
      </div>

      {/* 子フラグメント */}
      {showChildren && fragment.children && fragment.children.length > 0 && (
        <div className="mt-4 border-l-2 border-gray-600 pl-4 space-y-4">
          {fragment.children.map(child => (
            <FragmentCard
              key={child.id}
              fragment={child}
              onReply={onReply}
              replyTo={replyTo}
              setReplyTo={setReplyTo}
              loading={loading}
            />
          ))}
        </div>
      )}

      {/* 返信フォーム - このツリー内の返信対象の最下部に配置 */}
      {replyTo && isReplyTargetInThisTree(replyTo, fragment) && (
        <div className="mt-4 border-l-2 border-gray-600 pl-4">
          <FragmentForm
            onSubmit={onReply}
            loading={loading}
            placeholder="返信を書く..."
          />
        </div>
      )}
    </div>
  );
}
