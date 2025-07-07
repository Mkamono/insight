import { db, documents, documentTags, fragmentDocuments, tags, fragments, type Document, type NewDocument, type Tag, type Fragment } from '../db/index.js';
import { eq, sql, like, or, and, inArray } from 'drizzle-orm';
import { join, dirname } from 'path';
import { fileURLToPath } from 'url';
import { writeFileSync, mkdirSync } from 'fs';
import { handleError, InsightError } from '../utils/errorHandler.js';

export interface CreateDocumentInput extends Omit<NewDocument, 'id' | 'createdAt' | 'updatedAt'> {
  fragmentIds?: number[];
}

export interface UpdateDocumentInput extends Partial<Omit<NewDocument, 'id' | 'createdAt' | 'updatedAt'>> {
  fragmentIds?: number[];
}

export async function createDocument(input: CreateDocumentInput): Promise<Document> {
  try {
    const { fragmentIds, ...documentData } = input;
    
    const [result] = await db
      .insert(documents)
      .values(documentData)
      .returning();
    
    // フラグメントをリンク
    if (fragmentIds && fragmentIds.length > 0) {
      for (const fragmentId of fragmentIds) {
        await linkDocumentToFragment(result.id, fragmentId);
      }
    }
    
    return result;
  } catch (error) {
    throw handleError(error, 'createDocument');
  }
}

export async function findDocumentById(id: number): Promise<Document | null> {
  const result = await db
    .select()
    .from(documents)
    .where(eq(documents.id, id))
    .limit(1);
  return result[0] || null;
}

export async function findDocumentByTitle(title: string): Promise<Document | null> {
  const result = await db
    .select()
    .from(documents)
    .where(eq(documents.title, title))
    .limit(1);
  return result[0] || null;
}

export async function findAllDocuments(): Promise<Document[]> {
  return await db.select().from(documents);
}

export async function updateDocument(id: number, input: UpdateDocumentInput): Promise<Document | null> {
  const { fragmentIds, ...updateData } = input;
  
  const [result] = await db
    .update(documents)
    .set({ ...updateData, updatedAt: new Date() })
    .where(eq(documents.id, id))
    .returning();
  
  // フラグメントをリンク
  if (result && fragmentIds && fragmentIds.length > 0) {
    for (const fragmentId of fragmentIds) {
      await linkDocumentToFragment(result.id, fragmentId);
    }
  }
  
  return result || null;
}

export async function deleteDocument(id: number): Promise<boolean> {
  const result = await db
    .delete(documents)
    .where(eq(documents.id, id));
  return result.rowsAffected > 0;
}

export async function linkDocumentToFragment(documentId: number, fragmentId: number): Promise<void> {
  await db
    .insert(fragmentDocuments)
    .values({ documentId, fragmentId })
    .onConflictDoNothing();
}

export async function unlinkDocumentFromFragment(documentId: number, fragmentId: number): Promise<void> {
  await db
    .delete(fragmentDocuments)
    .where(
      sql`document_id = ${documentId} AND fragment_id = ${fragmentId}`
    );
}

export async function linkDocumentToTag(documentId: number, tagId: number): Promise<void> {
  await db
    .insert(documentTags)
    .values({ documentId, tagId })
    .onConflictDoNothing();
}

export async function unlinkDocumentFromTag(documentId: number, tagId: number): Promise<void> {
  await db
    .delete(documentTags)
    .where(
      sql`document_id = ${documentId} AND tag_id = ${tagId}`
    );
}

export async function getTagsByDocumentId(documentId: number): Promise<Tag[]> {
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

export async function getFragmentsByDocumentId(documentId: number): Promise<Fragment[]> {
  const result = await db
    .select({
      id: fragments.id,
      content: fragments.content,
      url: fragments.url,
      imagePath: fragments.imagePath,
      parentId: fragments.parentId,
      createdAt: fragments.createdAt,
      updatedAt: fragments.updatedAt,
    })
    .from(fragments)
    .innerJoin(fragmentDocuments, eq(fragments.id, fragmentDocuments.fragmentId))
    .where(eq(fragmentDocuments.documentId, documentId));
  return result;
}

export async function searchDocuments(options: {
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

function getDocumentsDir(): string {
  const __filename = fileURLToPath(import.meta.url);
  const __dirname = dirname(__filename);
  const projectRoot = join(__dirname, '../../..');
  return join(projectRoot, 'knowledge', 'documents');
}

function sanitizeFilename(title: string): string {
  // ファイル名に使えない文字を置換
  return title
    .replace(/[<>:"/\\|?*]/g, '_')  // 特殊文字を _ に置換
    .replace(/\s+/g, '_')          // 空白を _ に置換
    .substring(0, 100);            // 長すぎる場合は切り詰め
}

export async function generateMarkdownFile(document: Document): Promise<string> {
  try {
    // ドキュメントのタグを取得
    const documentTags = await getTagsByDocumentId(document.id);
    
    // ドキュメントのフラグメントを取得
    const documentFragments = await getFragmentsByDocumentId(document.id);

    // Markdownコンテンツを生成
    const markdown = generateMarkdownContent(document, documentTags, documentFragments);
    
    // ファイル名を生成
    const filename = `${sanitizeFilename(document.title)}.md`;
    
    // ディレクトリを確保
    const documentsDir = getDocumentsDir();
    mkdirSync(documentsDir, { recursive: true });
    
    // ファイルパスを生成
    const filePath = join(documentsDir, filename);
    
    // ファイルを書き込み
    writeFileSync(filePath, markdown, 'utf-8');
    
    console.log(`Markdownファイルを生成しました: ${filePath}`);
    return filePath;
  } catch (error) {
    throw handleError(error, 'generateMarkdownFile');
  }
}

function generateMarkdownContent(document: Document, tags: Tag[], fragments: Fragment[]): string {
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