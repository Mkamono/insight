import { NextResponse } from 'next/server';
import { TagService } from 'core';

const tagService = new TagService();

export async function GET() {
  try {
    const tags = await tagService.findAll();
    return NextResponse.json(tags);
  } catch (error) {
    console.error('Error fetching tags:', error);
    return NextResponse.json({ error: 'Failed to fetch tags' }, { status: 500 });
  }
}