import { db, tags, type Tag, type NewTag } from '../db/index.js';
import { eq, sql } from 'drizzle-orm';

export class TagService {
  async create(tag: NewTag): Promise<Tag> {
    const [result] = await db
      .insert(tags)
      .values(tag)
      .returning();
    return result;
  }

  async findById(id: number): Promise<Tag | null> {
    const result = await db
      .select()
      .from(tags)
      .where(eq(tags.id, id))
      .limit(1);
    return result[0] || null;
  }

  async findByName(name: string): Promise<Tag | null> {
    const result = await db
      .select()
      .from(tags)
      .where(eq(tags.name, name))
      .limit(1);
    return result[0] || null;
  }

  async findAll(): Promise<Tag[]> {
    return await db.select().from(tags);
  }

  async update(id: number, updates: Partial<NewTag>): Promise<Tag | null> {
    const [result] = await db
      .update(tags)
      .set({ ...updates, updatedAt: new Date() })
      .where(eq(tags.id, id))
      .returning();
    return result || null;
  }

  async delete(id: number): Promise<boolean> {
    const result = await db
      .delete(tags)
      .where(eq(tags.id, id));
    return result.rowsAffected > 0;
  }

  async findOrCreate(name: string): Promise<Tag> {
    const existing = await this.findByName(name);
    if (existing) {
      return existing;
    }
    return await this.create({ name });
  }
}