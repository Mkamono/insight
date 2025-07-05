import { NextResponse } from 'next/server';
import { AIService } from 'core';

export async function POST() {
  try {
    const aiService = new AIService();
    const documents = await aiService.processUnprocessedFragments();
    
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