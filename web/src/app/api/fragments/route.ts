import { NextRequest, NextResponse } from 'next/server';
import { findAllFragments, createFragment, findRootFragments, findFragmentWithChildren, findUnprocessedFragments } from 'core';

export async function GET(request: NextRequest) {
  try {
    const { searchParams } = new URL(request.url);
    const parentId = searchParams.get('parentId');
    const withChildren = searchParams.get('withChildren') === 'true';
    const rootOnly = searchParams.get('rootOnly') === 'true';
    const unprocessed = searchParams.get('unprocessed') === 'true';
    
    if (unprocessed) {
      const fragments = await findUnprocessedFragments();
      return NextResponse.json(fragments);
    }
    
    if (rootOnly) {
      const fragments = await findRootFragments();
      return NextResponse.json(fragments);
    }
    
    if (parentId && withChildren) {
      const fragment = await findFragmentWithChildren(parseInt(parentId));
      return NextResponse.json(fragment);
    }
    
    const fragments = await findAllFragments();
    return NextResponse.json(fragments);
  } catch (error) {
    console.error('Error fetching fragments:', error);
    return NextResponse.json({ error: 'Failed to fetch fragments' }, { status: 500 });
  }
}

export async function POST(request: NextRequest) {
  try {
    const body = await request.json();
    const { content, url, imagePath, parentId } = body;

    if (!content) {
      return NextResponse.json({ error: 'Content is required' }, { status: 400 });
    }

    const fragment = await createFragment({
      content,
      url: url || null,
      imagePath: imagePath || null,
      parentId: parentId || null,
    });

    return NextResponse.json(fragment);
  } catch (error) {
    console.error('Error creating fragment:', error);
    return NextResponse.json({ error: 'Failed to create fragment' }, { status: 500 });
  }
}