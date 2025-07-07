import { NextRequest, NextResponse } from 'next/server';
import { findFragmentById, findFragmentsByParentId, findFragmentHierarchy } from 'core';

export async function GET(request: NextRequest, { params }: { params: Promise<{ id: string }> }) {
  try {
    const { id: idString } = await params;
    const id = parseInt(idString);
    if (isNaN(id)) {
      return NextResponse.json({ error: 'Invalid fragment ID' }, { status: 400 });
    }

    const { searchParams } = new URL(request.url);
    const withChildren = searchParams.get('withChildren') === 'true';
    const withHierarchy = searchParams.get('withHierarchy') === 'true';

    const fragment = await findFragmentById(id);
    if (!fragment) {
      return NextResponse.json({ error: 'Fragment not found' }, { status: 404 });
    }

    if (withHierarchy) {
      const hierarchy = await findFragmentHierarchy(id);
      return NextResponse.json({ ...fragment, hierarchy });
    }

    if (withChildren) {
      const children = await findFragmentsByParentId(id);
      return NextResponse.json({ ...fragment, children });
    }

    return NextResponse.json(fragment);
  } catch (error) {
    console.error('Error fetching fragment:', error);
    return NextResponse.json({ error: 'Failed to fetch fragment' }, { status: 500 });
  }
}