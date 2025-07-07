import { db, questions, type Question, type NewQuestion } from '../db/index.js';
import { eq, and } from 'drizzle-orm';

export async function createQuestion(question: NewQuestion): Promise<Question> {
  const result = await db
    .insert(questions)
    .values(question)
    .returning();
  return result[0];
}

export async function findQuestionById(id: number): Promise<Question | null> {
  const result = await db
    .select()
    .from(questions)
    .where(eq(questions.id, id))
    .limit(1);
  return result[0] || null;
}

export async function findAllQuestions(): Promise<Question[]> {
  return await db.select().from(questions);
}

export async function findQuestionsByStatus(status: string): Promise<Question[]> {
  return await db
    .select()
    .from(questions)
    .where(eq(questions.status, status));
}

export async function findPendingQuestions(): Promise<Question[]> {
  return await findQuestionsByStatus('pending');
}

export async function updateQuestionStatus(id: number, status: string): Promise<Question | null> {
  const [result] = await db
    .update(questions)
    .set({ status, updatedAt: new Date() })
    .where(eq(questions.id, id))
    .returning();
  return result || null;
}

export async function markQuestionAsAnswered(id: number): Promise<Question | null> {
  return await updateQuestionStatus(id, 'answered');
}

export async function markQuestionAsDismissed(id: number): Promise<Question | null> {
  return await updateQuestionStatus(id, 'dismissed');
}

export async function deleteQuestion(id: number): Promise<boolean> {
  const result = await db
    .delete(questions)
    .where(eq(questions.id, id));
  return result.rowsAffected > 0;
}

export async function findQuestionsByDocument(documentId: number): Promise<Question[]> {
  return await db
    .select()
    .from(questions)
    .where(eq(questions.documentId, documentId));
}

export async function findQuestionsByFragment(fragmentId: number): Promise<Question[]> {
  return await db
    .select()
    .from(questions)
    .where(eq(questions.fragmentId, fragmentId));
}