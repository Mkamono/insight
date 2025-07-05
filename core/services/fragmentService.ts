import { db, fragments, type Fragment, type NewFragment } from '../db/index.js';
import { eq, sql } from 'drizzle-orm';

export class FragmentService {
  async create(fragment: NewFragment): Promise<Fragment> {
    const result = await db
      .insert(fragments)
      .values(fragment)
      .returning();
    return result[0];
  }

  async findById(id: number): Promise<Fragment | null> {
    const result = await db
      .select()
      .from(fragments)
      .where(eq(fragments.id, id))
      .limit(1);
    return result[0] || null;
  }

  async findAll(): Promise<Fragment[]> {
    return await db.select().from(fragments);
  }

  async update(id: number, updates: Partial<NewFragment>): Promise<Fragment | null> {
    const [result] = await db
      .update(fragments)
      .set({ ...updates, updatedAt: new Date() })
      .where(eq(fragments.id, id))
      .returning();
    return result || null;
  }

  async delete(id: number): Promise<boolean> {
    const result = await db
      .delete(fragments)
      .where(eq(fragments.id, id));
    return result.rowsAffected > 0;
  }

  async findByParentId(parentId: number): Promise<Fragment[]> {
    return await db
      .select()
      .from(fragments)
      .where(eq(fragments.parentId, parentId));
  }

  async markAsProcessed(id: number): Promise<Fragment | null> {
    return await this.update(id, { processed: true });
  }
}