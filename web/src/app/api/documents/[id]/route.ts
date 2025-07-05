import { NextRequest, NextResponse } from 'next/server';
import { DocumentService } from 'core';

const documentService = new DocumentService();

export async function GET(request: NextRequest, { params }: { params: Promise<{ id: string }> }) {
  try {
    const { id: idString } = await params;
    const id = parseInt(idString);
    if (isNaN(id)) {
      return NextResponse.json({ error: 'Invalid document ID' }, { status: 400 });
    }

    const document = await documentService.findById(id);
    
    if (!document) {
      return NextResponse.json({ error: 'Document not found' }, { status: 404 });
    }

    // ドキュメントに関連するタグとフラグメントを取得
    const [tags, fragments] = await Promise.all([
      documentService.getTagsByDocumentId(id),
      documentService.getFragmentsByDocumentId(id)
    ]);

    return NextResponse.json({
      ...document,
      tags,
      fragments
    });
  } catch (error) {
    console.error('Error fetching document:', error);
    return NextResponse.json({ error: 'Failed to fetch document' }, { status: 500 });
  }
}