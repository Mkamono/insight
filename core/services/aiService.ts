import { generateObject, generateText, LanguageModel, tool } from 'ai';
import { z } from 'zod';
import type { Document, Fragment } from '../db/index.js';
import { getAiModel } from './aimodel.js';
import { DocumentService } from './documentService.js';
import { FragmentService } from './fragmentService.js';
import { TagService } from './tagService.js';
export class AIService {
  private model: LanguageModel;
  private fragmentService: FragmentService;
  private documentService: DocumentService;
  private tagService: TagService;

  constructor() {
    this.model = getAiModel();
    this.fragmentService = new FragmentService();
    this.documentService = new DocumentService();
    this.tagService = new TagService();
  }

  async generateDocumentFromFragment(fragmentId: number): Promise<Document> {
    const fragment = await this.fragmentService.findById(fragmentId);
    if (!fragment) {
      throw new Error(`Fragment with id ${fragmentId} not found`);
    }

    // バッチ処理を利用して単一フラグメントを処理
    const documents = await this.processBatchFragments([fragment]);

    if (documents.length === 0) {
      throw new Error('Failed to generate document from fragment');
    }

    return documents[0];
  }


  async summarizeContent(content: string): Promise<string> {
    const prompt = `
以下の内容を簡潔に要約してください：

${content}

要件:
- 3-5文で要約してください
- 重要なポイントを含めてください
- 日本語で出力してください
`;

    const result = await generateText({
      model: this.model,
      prompt,
    });

    return result.text;
  }

  async generateQuestionsFromFragment(fragmentId: number): Promise<string[]> {
    const fragment = await this.fragmentService.findById(fragmentId);
    if (!fragment) {
      throw new Error(`Fragment with id ${fragmentId} not found`);
    }

    const prompt = `
以下のフラグメントを基に、理解を深めるための質問を3-5個生成してください：

フラグメント内容: ${fragment.content}
${fragment.url ? `URL: ${fragment.url}` : ''}

要件:
- 内容の理解を深める質問を生成してください
- 実用的で具体的な質問にしてください
- 日本語で出力してください
`;

    const result = await generateObject({
      model: this.model,
      prompt,
      schema: z.object({
        questions: z.array(z.string()).describe('生成された質問のリスト'),
      }),
    });

    return result.object.questions;
  }

  private createDocumentTool = tool({
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
        const document = await this.documentService.create({ title, content, summary });

        // タグをリンク
        for (const tagName of tags) {
          const tag = await this.tagService.findOrCreate(tagName);
          await this.documentService.linkToTag(document.id, tag.id);
        }

        // フラグメントをリンクして処理済みにマーク
        for (const fragmentId of fragmentIds) {
          await this.documentService.linkToFragment(document.id, fragmentId);
          await this.fragmentService.markAsProcessed(fragmentId);
        }

        // Markdownファイルを生成
        await this.documentService.generateMarkdownFile(document);

        return { success: true, documentId: document.id, title: document.title };
      } catch (error) {
        return { success: false, error: error instanceof Error ? error.message : 'Unknown error' };
      }
    },
  });

  private updateDocumentTool = tool({
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
        const document = await this.documentService.update(documentId, { title, content, summary });

        if (!document) {
          return { success: false, error: 'Document not found or failed to update' };
        }

        // タグをリンク
        for (const tagName of tags) {
          const tag = await this.tagService.findOrCreate(tagName);
          await this.documentService.linkToTag(document.id, tag.id);
        }

        // フラグメントをリンクして処理済みにマーク
        for (const fragmentId of fragmentIds) {
          await this.documentService.linkToFragment(document.id, fragmentId);
          await this.fragmentService.markAsProcessed(fragmentId);
        }

        // Markdownファイルを生成（更新の場合は上書き）
        await this.documentService.generateMarkdownFile(document);

        return { success: true, documentId: document.id, title: document.title, updated: true };
      } catch (error) {
        return { success: false, error: error instanceof Error ? error.message : 'Unknown error' };
      }
    },
  });

  private getExistingDocumentsInfoTool = tool({
    description: 'Get basic information about existing documents to help with update decisions',
    parameters: z.object({}),
    execute: async () => {
      try {
        const documents = await this.documentService.findAll();
        const documentsInfo = documents.map(doc => ({
          id: doc.id,
          title: doc.title,
          summary: doc.summary,
          contentLength: doc.content.length,
        }));
        return { success: true, documents: documentsInfo };
      } catch (error) {
        return { success: false, error: error instanceof Error ? error.message : 'Unknown error' };
      }
    },
  });

  private getDocumentDetailTool = tool({
    description: 'Get detailed content of a specific document to make update decisions',
    parameters: z.object({
      documentId: z.number().describe('Document ID to get details for'),
    }),
    execute: async ({ documentId }) => {
      try {
        const document = await this.documentService.findById(documentId);
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

  private linkFragmentToDocumentTool = tool({
    description: 'Link a fragment to a document',
    parameters: z.object({
      fragmentId: z.number().describe('Fragment ID to link'),
      documentId: z.number().describe('Document ID to link to'),
    }),
    execute: async ({ fragmentId, documentId }) => {
      try {
        await this.documentService.linkToFragment(documentId, fragmentId);
        await this.fragmentService.markAsProcessed(fragmentId);

        return { success: true, documentId, fragmentId };
      } catch (error) {
        return { success: false, error: error instanceof Error ? error.message : 'Unknown error' };
      }
    },
  });

  async processUnprocessedFragments(): Promise<Document[]> {
    const fragments = await this.fragmentService.findAll();
    const unprocessedFragments = fragments.filter(f => !f.processed);

    if (unprocessedFragments.length === 0) {
      console.log('処理するフラグメントがありません');
      return [];
    }

    try {
      return await this.processBatchFragments(unprocessedFragments);
    } catch (error) {
      console.error('処理中にエラーが発生しました:', error);
      // レート制限エラーの場合
      if (error instanceof Error && (
        error.message.includes('rate limit') ||
        error.message.includes('quota') ||
        error.message.includes('429')
      )) {
        console.log('レート制限が検出されました。少し時間をおいてから再実行してください。');
      }
      return [];
    }
  }

  private async processBatchFragments(fragments: Fragment[]): Promise<Document[]> {
    const existingDocuments = await this.documentService.findAll();

    // 既存ドキュメントの情報を簡潔にまとめる
    const existingDocsInfo = existingDocuments.map(doc => ({
      id: doc.id,
      title: doc.title,
      summary: doc.summary
    }));

    const prompt = `
複数のフラグメントを効率的に処理して、適切にドキュメントを作成・更新してください。

未処理フラグメント:
${fragments.map(f => `ID: ${f.id} - ${f.content}`).join('\n')}

既存ドキュメント:
${existingDocsInfo.length > 0 ? existingDocsInfo.map(doc =>
      `ID: ${doc.id}, タイトル: ${doc.title}, 要約: ${doc.summary}`
    ).join('\n') : '既存ドキュメントはありません'}

指示:
1. 各フラグメントについて、既存ドキュメントとの関連性を判断してください
2. 関連性がある場合は既存ドキュメントの更新を、ない場合は新規作成を指示してください
3. 複数のフラグメントが同じ新規ドキュメントに統合可能な場合は、統合してください
4. 適切なタグも提案してください
5. **重要**: createDocumentまたはupdateDocumentを呼び出す際、fragmentIdsパラメータに関連するフラグメントIDを必ず含めてください
6. **重要**: contentパラメータはMarkdown形式で記述してください（見出し、リスト、強調などを使用）

処理例:
- createDocument(title: "Python概要", content: "# Python\n\n## 特徴\n- 可読性に優れている\n- 動的型付け\n\n## 用途\nWebアプリケーション開発等", summary: "...", tags: [...], fragmentIds: [1, 2])

利用可能なツール:
- getExistingDocumentsInfo: 既存ドキュメント一覧を取得
- getDocumentDetail: 特定ドキュメントの詳細を取得
- createDocument: 新しいドキュメントを作成（content=Markdown形式、fragmentIds必須）
- updateDocument: 既存ドキュメントを更新（content=Markdown形式、fragmentIds必須）
`;

    await generateText({
      model: this.model,
      prompt,
      tools: {
        getExistingDocumentsInfo: this.getExistingDocumentsInfoTool,
        getDocumentDetail: this.getDocumentDetailTool,
        createDocument: this.createDocumentTool,
        updateDocument: this.updateDocumentTool,
      },
      maxSteps: 20,
    });

    // 処理されたドキュメントを取得
    const updatedDocuments = await this.documentService.findAll();
    const processedDocuments = updatedDocuments.filter(doc => {
      const existingDoc = existingDocuments.find(existing => existing.id === doc.id);
      if (!existingDoc) {
        // 新規ドキュメント
        return true;
      }
      // 更新されたドキュメント（時間比較）
      return existingDoc.updatedAt && doc.updatedAt &&
        existingDoc.updatedAt.getTime() !== doc.updatedAt.getTime();
    });

    return processedDocuments;
  }

}
