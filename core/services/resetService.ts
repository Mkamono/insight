import { db } from '../db/index.js';
import { sql } from 'drizzle-orm';
import { migrate } from 'drizzle-orm/libsql/migrator';
import { join, dirname } from 'path';
import { fileURLToPath } from 'url';
import { rmSync, existsSync, mkdirSync } from 'fs';

export async function resetDatabase(): Promise<void> {
  try {
    // 1. 全てのテーブルを削除（DBファイルは保持）
    await db.run(sql`DROP TABLE IF EXISTS fragment_documents`);
    await db.run(sql`DROP TABLE IF EXISTS document_tags`);
    await db.run(sql`DROP TABLE IF EXISTS documents`);
    await db.run(sql`DROP TABLE IF EXISTS fragments`);
    await db.run(sql`DROP TABLE IF EXISTS tags`);
    console.log('既存のテーブルを削除しました');
    
    // 2. マイグレーションを実行してテーブルを再作成
    await migrate(db, { migrationsFolder: './drizzle' });
    console.log('データベースを再作成しました');
    
  } catch (error) {
    console.error('データベースリセット中にエラーが発生しました:', error);
    throw error;
  }
}

export async function resetDocuments(): Promise<void> {
  try {
    // ドキュメントディレクトリ内のMarkdownファイルを削除
    const __filename = fileURLToPath(import.meta.url);
    const __dirname = dirname(__filename);
    const projectRoot = join(__dirname, '../../..');
    const documentsDir = join(projectRoot, 'knowledge', 'documents');
    
    if (existsSync(documentsDir)) {
      rmSync(documentsDir, { recursive: true, force: true });
      console.log('既存のドキュメントファイルを削除しました');
    }
    
    // ディレクトリを再作成
    mkdirSync(documentsDir, { recursive: true });
    console.log('ドキュメントディレクトリを再作成しました');
    
  } catch (error) {
    console.error('ドキュメントリセット中にエラーが発生しました:', error);
    throw error;
  }
}

export async function resetAll(): Promise<void> {
  try {
    console.log('システム全体をリセット中...');
    
    // 1. ドキュメントファイルを削除
    await resetDocuments();
    
    // 2. データベースをリセット
    await resetDatabase();
    
    console.log('システムリセットが完了しました');
    
  } catch (error) {
    console.error('システムリセット中にエラーが発生しました:', error);
    throw error;
  }
}