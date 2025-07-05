import { createClient } from '@libsql/client';
import { drizzle } from 'drizzle-orm/libsql';
import { join, dirname } from 'path';
import { fileURLToPath } from 'url';
import * as schema from './schema.js';

// プロジェクトルートのknowledgeディレクトリへの絶対パスを取得
const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);
const projectRoot = join(__dirname, '../../..');
const dbPath = join(projectRoot, 'knowledge', 'data.db');

const client = createClient({
  url: process.env.DATABASE_URL || `file:${dbPath}`,
});

export const db = drizzle(client, { schema });
export * from './schema.js';
