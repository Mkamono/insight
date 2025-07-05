import { db, questions, type Question, type NewQuestion } from '../db/index.js';
import { eq, sql } from 'drizzle-orm';

export class QuestionService {
  async create(question: NewQuestion): Promise<Question> {
    const [result] = await db
      .insert(questions)
      .values(question)
      .returning();
    return result;
  }

  async findById(id: number): Promise<Question | null> {
    const result = await db
      .select()
      .from(questions)
      .where(eq(questions.id, id))
      .limit(1);
    return result[0] || null;
  }

  async findAll(): Promise<Question[]> {
    return await db.select().from(questions);
  }

  async update(id: number, updates: Partial<NewQuestion>): Promise<Question | null> {
    const [result] = await db
      .update(questions)
      .set({ ...updates, updatedAt: new Date() })
      .where(eq(questions.id, id))
      .returning();
    return result || null;
  }

  async delete(id: number): Promise<boolean> {
    const result = await db
      .delete(questions)
      .where(eq(questions.id, id));
    return result.rowsAffected > 0;
  }
}