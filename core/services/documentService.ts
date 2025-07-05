import { db, documents, documentTags, fragmentDocuments, tags, fragments, type Document, type NewDocument, type Tag, type Fragment } from '../db/index.js';
import { eq, sql, like, or, and, inArray } from 'drizzle-orm';
import { join, dirname } from 'path';
import { fileURLToPath } from 'url';
import { writeFileSync, mkdirSync } from 'fs';
import { handleError, InsightError } from '../utils/errorHandler.js';

export class DocumentService {
  async create(document: NewDocument): Promise<Document> {
    try {
      const [result] = await db
        .insert(documents)
        .values(document)
        .returning();
      return result;
    } catch (error) {
      throw handleError(error, 'DocumentService.create');
    }
  }

  async findById(id: number): Promise<Document | null> {
    const result = await db
      .select()
      .from(documents)
      .where(eq(documents.id, id))
      .limit(1);
    return result[0] || null;
  }

  async findByTitle(title: string): Promise<Document | null> {
    const result = await db
      .select()
      .from(documents)
      .where(eq(documents.title, title))
      .limit(1);
    return result[0] || null;
  }

  async findAll(): Promise<Document[]> {
    return await db.select().from(documents);
  }

  async update(id: number, updates: Partial<NewDocument>): Promise<Document | null> {
    const [result] = await db
      .update(documents)
      .set({ ...updates, updatedAt: new Date() })
      .where(eq(documents.id, id))
      .returning();
    return result || null;
  }

  async delete(id: number): Promise<boolean> {
    const result = await db
      .delete(documents)
      .where(eq(documents.id, id));
    return result.rowsAffected > 0;
  }

  async linkToFragment(documentId: number, fragmentId: number): Promise<void> {
    await db
      .insert(fragmentDocuments)
      .values({ documentId, fragmentId })
      .onConflictDoNothing();
  }

  async unlinkFromFragment(documentId: number, fragmentId: number): Promise<void> {
    await db
      .delete(fragmentDocuments)
      .where(
        sql`document_id = ${documentId} AND fragment_id = ${fragmentId}`
      );
  }

  async linkToTag(documentId: number, tagId: number): Promise<void> {
    await db
      .insert(documentTags)
      .values({ documentId, tagId })
      .onConflictDoNothing();
  }

  async unlinkFromTag(documentId: number, tagId: number): Promise<void> {
    await db
      .delete(documentTags)
      .where(
        sql`document_id = ${documentId} AND tag_id = ${tagId}`
      );
  }

  async getTagsByDocumentId(documentId: number): Promise<Tag[]> {
    const result = await db
      .select({
        id: tags.id,
        name: tags.name,
        createdAt: tags.createdAt,
        updatedAt: tags.updatedAt,
      })
      .from(tags)
      .innerJoin(documentTags, eq(tags.id, documentTags.tagId))
      .where(eq(documentTags.documentId, documentId));
    return result;
  }

  async getFragmentsByDocumentId(documentId: number): Promise<Fragment[]> {
    const result = await db
      .select({
        id: fragments.id,
        content: fragments.content,
        url: fragments.url,
        imagePath: fragments.imagePath,
        processed: fragments.processed,
        parentId: fragments.parentId,
        createdAt: fragments.createdAt,
        updatedAt: fragments.updatedAt,
      })
      .from(fragments)
      .innerJoin(fragmentDocuments, eq(fragments.id, fragmentDocuments.fragmentId))
      .where(eq(fragmentDocuments.documentId, documentId));
    return result;
  }

  async searchDocuments(options: {
    query?: string;
    tagIds?: number[];
  }): Promise<Document[]> {
    // 条件配列を構築
    const conditions = [];

    // テキスト検索条件
    if (options.query && options.query.trim()) {
      const searchTerm = `%${options.query.trim()}%`;
      conditions.push(
        or(
          like(documents.title, searchTerm),
          like(documents.content, searchTerm),
          like(documents.summary, searchTerm)
        )
      );
    }

    // タグフィルター条件
    if (options.tagIds && options.tagIds.length > 0) {
      // タグIDに一致するドキュメントIDを取得
      const docsWithTags = await db
        .select({ documentId: documentTags.documentId })
        .from(documentTags)
        .where(inArray(documentTags.tagId, options.tagIds));
      
      if (docsWithTags.length > 0) {
        const documentIds = docsWithTags.map(doc => doc.documentId);
        conditions.push(inArray(documents.id, documentIds));
      } else {
        // 指定されたタグを持つドキュメントが存在しない場合は空の結果を返す
        return [];
      }
    }

    // クエリを実行
    if (conditions.length > 0) {
      return await db
        .select()
        .from(documents)
        .where(and(...conditions));
    } else {
      return await db.select().from(documents);
    }
  }

  private getDocumentsDir(): string {
    const __filename = fileURLToPath(import.meta.url);
    const __dirname = dirname(__filename);
    const projectRoot = join(__dirname, '../../..');
    return join(projectRoot, 'knowledge', 'documents');
  }

  private sanitizeFilename(title: string): string {
    // ファイル名に使えない文字を置換
    return title
      .replace(/[<>:"/\\|?*]/g, '_')  // 特殊文字を _ に置換
      .replace(/\s+/g, '_')          // 空白を _ に置換
      .substring(0, 100);            // 長すぎる場合は切り詰め
  }

  async generateMarkdownFile(document: Document): Promise<string> {
    try {
      // ドキュメントのタグを取得
      const documentTags = await this.getTagsByDocumentId(document.id);
      
      // ドキュメントのフラグメントを取得
      const documentFragments = await this.getFragmentsByDocumentId(document.id);

      // Markdownコンテンツを生成
      const markdown = this.generateMarkdownContent(document, documentTags, documentFragments);
      
      // ファイル名を生成
      const filename = `${this.sanitizeFilename(document.title)}.md`;
      
      // ディレクトリを確保
      const documentsDir = this.getDocumentsDir();
      mkdirSync(documentsDir, { recursive: true });
      
      // ファイルパスを生成
      const filePath = join(documentsDir, filename);
      
      // ファイルを書き込み
      writeFileSync(filePath, markdown, 'utf-8');
      
      console.log(`Markdownファイルを生成しました: ${filePath}`);
      return filePath;
    } catch (error) {
      throw handleError(error, 'DocumentService.generateMarkdownFile');
    }
  }

  private generateMarkdownContent(document: Document, tags: Tag[], fragments: Fragment[]): string {
    const formatDate = (date: Date | null) => date ? date.toISOString().split('T')[0] : 'N/A';
    
    return `${document.content}

---

## メタデータ

| 項目 | 内容 |
|------|------|
| **要約** | ${document.summary} |
| **タグ** | ${tags.length > 0 ? tags.map(tag => tag.name).join(', ') : '(なし)'} |
| **作成日** | ${formatDate(document.createdAt)} |
| **更新日** | ${formatDate(document.updatedAt)} |
| **ドキュメントID** | ${document.id} |

## 参考フラグメント

| ID | 内容 | URL | 画像 |
|----|------|-----|------|
${fragments.length > 0 ? fragments.map(fragment => 
  `| ${fragment.id} | ${fragment.content} | ${fragment.url || '-'} | ${fragment.imagePath || '-'} |`
).join('\n') : '| - | フラグメントなし | - | - |'}
`;
  }
}