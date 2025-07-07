import { tool } from 'ai';
import { z } from 'zod';
import { createDocument, findDocumentById, findAllDocuments, generateMarkdownFile, linkDocumentToTag, updateDocument } from './documentService.js';
import { findOrCreateTag } from './tagService.js';
import { createQuestion } from './questionService.js';

export const createDocumentTool = tool({
  description: 'Create a new document with title, content, summary, and tags, and link fragments',
  parameters: z.object({
    title: z.string().describe('Document title'),
    content: z.string().describe('Document content'),
    summary: z.string().describe('Document summary'),
    tags: z.array(z.string()).describe('Array of tag names'),
    fragmentIds: z.array(z.number()).describe('Array of fragment IDs to link to this document'),
  }),
  execute: async ({ title, content, summary, tags, fragmentIds }) => {
    try {
      const document = await createDocument({ 
        title, 
        content, 
        summary, 
        fragmentIds 
      });

      // タグをリンク
      for (const tagName of tags) {
        const tag = await findOrCreateTag(tagName);
        await linkDocumentToTag(document.id, tag.id);
      }

      // Markdownファイルを生成
      await generateMarkdownFile(document);

      return { success: true, documentId: document.id, title: document.title };
    } catch (error) {
      return { success: false, error: error instanceof Error ? error.message : 'Unknown error' };
    }
  },
});

export const updateDocumentTool = tool({
  description: 'Update an existing document by ID and link fragments',
  parameters: z.object({
    documentId: z.number().describe('Document ID to update'),
    title: z.string().describe('New document title'),
    content: z.string().describe('New document content'),
    summary: z.string().describe('New document summary'),
    tags: z.array(z.string()).describe('Array of tag names'),
    fragmentIds: z.array(z.number()).describe('Array of fragment IDs to link to this document'),
  }),
  execute: async ({ documentId, title, content, summary, tags, fragmentIds }) => {
    try {
      const document = await updateDocument(documentId, { 
        title, 
        content, 
        summary, 
        fragmentIds 
      });

      if (!document) {
        return { success: false, error: 'Document not found or failed to update' };
      }

      // タグをリンク
      for (const tagName of tags) {
        const tag = await findOrCreateTag(tagName);
        await linkDocumentToTag(document.id, tag.id);
      }

      // Markdownファイルを生成（更新の場合は上書き）
      await generateMarkdownFile(document);

      return { success: true, documentId: document.id, title: document.title, updated: true };
    } catch (error) {
      return { success: false, error: error instanceof Error ? error.message : 'Unknown error' };
    }
  },
});


export const getDocumentDetailTool = tool({
  description: 'Get detailed content of a specific document to make update decisions',
  parameters: z.object({
    documentId: z.number().describe('Document ID to get details for'),
  }),
  execute: async ({ documentId }) => {
    try {
      const document = await findDocumentById(documentId);
      if (!document) {
        return { success: false, error: 'Document not found' };
      }

      return {
        success: true,
        document: {
          id: document.id,
          title: document.title,
          content: document.content,
          summary: document.summary,
          createdAt: document.createdAt,
          updatedAt: document.updatedAt,
        }
      };
    } catch (error) {
      return { success: false, error: error instanceof Error ? error.message : 'Unknown error' };
    }
  },
});

export const createQuestionTool = tool({
  description: 'Create a question when information is missing or incomplete to guide user to provide more context',
  parameters: z.object({
    content: z.string().describe('The question to ask the user'),
    context: z.string().describe('Context explaining why this question is needed'),
    documentId: z.number().optional().describe('Related document ID if applicable'),
    fragmentId: z.number().optional().describe('Related fragment ID if applicable'),
  }),
  execute: async ({ content, context, documentId, fragmentId }) => {
    try {
      const question = await createQuestion({
        content,
        context,
        documentId: documentId || null,
        fragmentId: fragmentId || null,
        status: 'pending',
      });

      return { 
        success: true, 
        questionId: question.id, 
        content: question.content,
        message: 'Question created to gather more information from user'
      };
    } catch (error) {
      return { success: false, error: error instanceof Error ? error.message : 'Unknown error' };
    }
  },
});

