import { db, tags, type Tag, type NewTag } from '../db/index.js';
import { eq, sql } from 'drizzle-orm';

export async function createTag(tag: NewTag): Promise<Tag> {
  const [result] = await db
    .insert(tags)
    .values(tag)
    .returning();
  return result;
}

export async function findTagById(id: number): Promise<Tag | null> {
  const result = await db
    .select()
    .from(tags)
    .where(eq(tags.id, id))
    .limit(1);
  return result[0] || null;
}

export async function findTagByName(name: string): Promise<Tag | null> {
  const result = await db
    .select()
    .from(tags)
    .where(eq(tags.name, name))
    .limit(1);
  return result[0] || null;
}

export async function findAllTags(): Promise<Tag[]> {
  return await db.select().from(tags);
}

export async function updateTag(id: number, updates: Partial<NewTag>): Promise<Tag | null> {
  const [result] = await db
    .update(tags)
    .set({ ...updates, updatedAt: new Date() })
    .where(eq(tags.id, id))
    .returning();
  return result || null;
}

export async function deleteTag(id: number): Promise<boolean> {
  const result = await db
    .delete(tags)
    .where(eq(tags.id, id));
  return result.rowsAffected > 0;
}

export async function findOrCreateTag(name: string): Promise<Tag> {
  const existing = await findTagByName(name);
  if (existing) {
    return existing;
  }
  return await createTag({ name });
}