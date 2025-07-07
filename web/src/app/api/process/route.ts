import { NextResponse } from 'next/server';
import { generateDocumentFromFragment, findUnprocessedFragments } from 'core';

export async function POST() {
  try {
    // 未処理フラグメントを取得
    const unprocessedFragments = await findUnprocessedFragments();
    
    if (unprocessedFragments.length === 0) {
      return NextResponse.json({ documents: [] });
    }
    
    // フラグメントIDを取得
    const fragmentIds = unprocessedFragments.map(f => f.id);
    
    // ドキュメントを生成
    const documents = await generateDocumentFromFragment(fragmentIds);
    
    return NextResponse.json({
      documents: documents.map(doc => ({
        id: doc.id,
        title: doc.title,
        summary: doc.summary,
      })),
    });
  } catch (error) {
    console.error('Error processing fragments:', error);
    return NextResponse.json({ error: 'Failed to process fragments' }, { status: 500 });
  }
}