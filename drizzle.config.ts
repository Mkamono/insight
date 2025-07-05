import type { Config } from 'drizzle-kit';
import { join } from 'path';

export default {
  schema: './core/db/schema.ts',
  out: './drizzle',
  dialect: 'sqlite',
  dbCredentials: {
    url: process.env.DATABASE_URL || `file:${join(process.cwd(), 'knowledge', 'data.db')}`,
  },
} satisfies Config;