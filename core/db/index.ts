import { createClient } from '@libsql/client';
import { drizzle } from 'drizzle-orm/libsql';
import { dirname, join } from 'path';
import { fileURLToPath } from 'url';
import { existsSync, mkdirSync } from 'fs';
import * as schema from './schema.js';

// プロジェクトルートのknowledgeディレクトリへの絶対パスを取得
const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);
const projectRoot = join(__dirname, '../../..');
const knowledgeDir = join(projectRoot, 'knowledge');
const dbPath = join(knowledgeDir, 'data.db');

// データベース接続を遅延初期化
function initializeDatabase() {
  // knowledgeディレクトリが存在しない場合は作成
  if (!existsSync(knowledgeDir)) {
    mkdirSync(knowledgeDir, { recursive: true });
  }
  
  const client = createClient({
    url: `file:${dbPath}`,
  });
  
  return drizzle(client, { schema });
}

export const db = initializeDatabase();
export * from './schema.js';
