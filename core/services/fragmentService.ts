import { db, fragments, fragmentDocuments, type Fragment, type NewFragment } from '../db/index.js';
import { eq, sql, notInArray } from 'drizzle-orm';

export async function createFragment(fragment: NewFragment): Promise<Fragment> {
  const result = await db
    .insert(fragments)
    .values(fragment)
    .returning();
  return result[0];
}

export async function findFragmentById(id: number): Promise<Fragment | null> {
  const result = await db
    .select()
    .from(fragments)
    .where(eq(fragments.id, id))
    .limit(1);
  return result[0] || null;
}

export async function findAllFragments(): Promise<Fragment[]> {
  return await db.select().from(fragments);
}

export async function updateFragment(id: number, updates: Partial<NewFragment>): Promise<Fragment | null> {
  const [result] = await db
    .update(fragments)
    .set({ ...updates, updatedAt: new Date() })
    .where(eq(fragments.id, id))
    .returning();
  return result || null;
}

export async function deleteFragment(id: number): Promise<boolean> {
  const result = await db
    .delete(fragments)
    .where(eq(fragments.id, id));
  return result.rowsAffected > 0;
}

export async function findFragmentsByParentId(parentId: number): Promise<Fragment[]> {
  return await db
    .select()
    .from(fragments)
    .where(eq(fragments.parentId, parentId));
}

export async function findRootFragments(): Promise<Fragment[]> {
  return await db
    .select()
    .from(fragments)
    .where(sql`${fragments.parentId} IS NULL`);
}

export async function findFragmentWithChildren(id: number): Promise<Fragment & { children: Fragment[] } | null> {
  const fragment = await findFragmentById(id);
  if (!fragment) return null;
  
  const children = await findFragmentsByParentId(id);
  return { ...fragment, children };
}

export async function findFragmentHierarchy(id: number): Promise<Fragment[]> {
  const hierarchy: Fragment[] = [];
  let currentId: number | null = id;
  
  while (currentId) {
    const fragment = await findFragmentById(currentId);
    if (!fragment) break;
    
    hierarchy.unshift(fragment);
    currentId = fragment.parentId;
  }
  
  return hierarchy;
}

export async function findUnprocessedFragments(): Promise<Fragment[]> {
  // フラグメント-ドキュメントリンクテーブルから、リンクされているフラグメントIDを取得
  const linkedFragmentIds = await db
    .selectDistinct({ fragmentId: fragmentDocuments.fragmentId })
    .from(fragmentDocuments);
  
  if (linkedFragmentIds.length === 0) {
    // リンクされているフラグメントがない場合は、全フラグメントが未処理
    return await findAllFragments();
  }
  
  const linkedIds = linkedFragmentIds.map(row => row.fragmentId);
  
  // リンクされていないフラグメントを取得
  return await db
    .select()
    .from(fragments)
    .where(notInArray(fragments.id, linkedIds));
}