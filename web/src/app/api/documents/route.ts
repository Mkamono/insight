import { NextResponse } from 'next/server';
import { findAllDocuments } from 'core';

export async function GET() {
  try {
    const documents = await findAllDocuments();
    return NextResponse.json(documents);
  } catch (error) {
    console.error('Error fetching documents:', error);
    return NextResponse.json({ error: 'Failed to fetch documents' }, { status: 500 });
  }
}