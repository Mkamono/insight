import { NextRequest, NextResponse } from 'next/server';
import { FragmentService } from 'core';

const fragmentService = new FragmentService();

export async function GET() {
  try {
    const fragments = await fragmentService.findAll();
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

    const fragment = await fragmentService.create({
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