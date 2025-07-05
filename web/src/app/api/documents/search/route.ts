import { NextRequest, NextResponse } from 'next/server';
import { DocumentService } from 'core';

const documentService = new DocumentService();

export async function GET(request: NextRequest) {
  try {
    const { searchParams } = new URL(request.url);
    const query = searchParams.get('q');
    const tagIdsParam = searchParams.get('tags');
    
    let tagIds: number[] = [];
    if (tagIdsParam) {
      tagIds = tagIdsParam.split(',').map(id => parseInt(id)).filter(id => !isNaN(id));
    }

    const documents = await documentService.searchDocuments({
      query: query || undefined,
      tagIds: tagIds.length > 0 ? tagIds : undefined,
    });

    return NextResponse.json(documents);
  } catch (error) {
    console.error('Error searching documents:', error);
    return NextResponse.json({ error: 'Failed to search documents' }, { status: 500 });
  }
}