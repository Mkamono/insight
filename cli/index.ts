#!/usr/bin/env node

import { Command } from 'commander';
import { 
  createFragment, findAllFragments, findFragmentById,
  createDocument, findAllDocuments, findDocumentById,
  createTag, findAllTags,
  generateDocumentFromFragment,
  resetAll, resetDatabase, resetDocuments
} from '../core/index.js';

const program = new Command();

program
  .name('insight')
  .description('AI知識ドキュメント化システム')
  .version('1.0.0');

// Fragment commands
const fragmentCmd = program
  .command('fragment')
  .description('フラグメント管理');

fragmentCmd
  .command('create')
  .description('新しいフラグメントを作成')
  .option('-c, --content <content>', 'フラグメントの内容')
  .option('-u, --url <url>', 'URL（オプション）')
  .option('-i, --image <path>', '画像パス（オプション）')
  .option('-p, --parent <id>', '親フラグメントID（オプション）')
  .action(async (options) => {
    try {
      const fragment = await createFragment({
        content: options.content,
        url: options.url || null,
        imagePath: options.image || null,
        parentId: options.parent ? parseInt(options.parent) : null,
      });
      console.log('フラグメントが作成されました:', fragment);
    } catch (error) {
      console.error('エラー:', error);
    }
  });

fragmentCmd
  .command('list')
  .description('全てのフラグメントを表示')
  .action(async () => {
    try {
      const fragments = await findAllFragments();
      console.log('フラグメント一覧:');
      fragments.forEach((f) => {
        console.log(`ID: ${f.id}, 内容: ${f.content.slice(0, 50)}...`);
      });
    } catch (error) {
      console.error('エラー:', error);
    }
  });

fragmentCmd
  .command('get <id>')
  .description('IDでフラグメントを取得')
  .action(async (id) => {
    try {
      const fragment = await findFragmentById(parseInt(id));
      if (fragment) {
        console.log('フラグメント:', fragment);
      } else {
        console.log('フラグメントが見つかりません');
      }
    } catch (error) {
      console.error('エラー:', error);
    }
  });

// Document commands
const documentCmd = program
  .command('document')
  .description('ドキュメント管理');

documentCmd
  .command('create')
  .description('新しいドキュメントを作成')
  .option('-t, --title <title>', 'ドキュメントのタイトル')
  .option('-c, --content <content>', 'ドキュメントの内容')
  .option('-s, --summary <summary>', 'ドキュメントの要約')
  .action(async (options) => {
    try {
      const document = await createDocument({
        title: options.title,
        content: options.content,
        summary: options.summary,
      });
      console.log('ドキュメントが作成されました:', document);
    } catch (error) {
      console.error('エラー:', error);
    }
  });

documentCmd
  .command('list')
  .description('全てのドキュメントを表示')
  .action(async () => {
    try {
      const documents = await findAllDocuments();
      console.log('ドキュメント一覧:');
      documents.forEach((d) => {
        console.log(`ID: ${d.id}, タイトル: ${d.title}`);
      });
    } catch (error) {
      console.error('エラー:', error);
    }
  });

documentCmd
  .command('get <id>')
  .description('IDでドキュメントを取得')
  .action(async (id) => {
    try {
      const document = await findDocumentById(parseInt(id));
      if (document) {
        console.log('ドキュメント:', document);
      } else {
        console.log('ドキュメントが見つかりません');
      }
    } catch (error) {
      console.error('エラー:', error);
    }
  });


// Tag commands
const tagCmd = program
  .command('tag')
  .description('タグ管理');

tagCmd
  .command('create')
  .description('新しいタグを作成')
  .option('-n, --name <name>', 'タグ名')
  .action(async (options) => {
    try {
      const tag = await createTag({
        name: options.name,
      });
      console.log('タグが作成されました:', tag);
    } catch (error) {
      console.error('エラー:', error);
    }
  });

tagCmd
  .command('list')
  .description('全てのタグを表示')
  .action(async () => {
    try {
      const tags = await findAllTags();
      console.log('タグ一覧:');
      tags.forEach((t) => {
        console.log(`ID: ${t.id}, 名前: ${t.name}`);
      });
    } catch (error) {
      console.error('エラー:', error);
    }
  });

// Database initialization
program
  .command('init')
  .description('データベースを初期化')
  .action(async () => {
    try {
      const { existsSync, mkdirSync } = await import('fs');
      const { join } = await import('path');
      
      // knowledgeディレクトリが存在しない場合は作成
      const knowledgeDir = join(process.cwd(), 'knowledge');
      if (!existsSync(knowledgeDir)) {
        mkdirSync(knowledgeDir, { recursive: true });
        console.log('knowledgeディレクトリを作成しました');
      }
      
      const { migrate } = await import('drizzle-orm/libsql/migrator');
      const { db } = await import('../core/db/index.js');
      
      await migrate(db, { migrationsFolder: './drizzle' });
      console.log('データベースが初期化されました');
    } catch (error) {
      console.error('データベース初期化エラー:', error);
    }
  });

// Reset commands
const resetCmd = program
  .command('reset')
  .description('リセット機能');

resetCmd
  .command('all')
  .description('データベースとドキュメントを全てリセット')
  .action(async () => {
    try {
      await resetAll();
    } catch (error) {
      console.error('エラー:', error);
    }
  });

resetCmd
  .command('database')
  .description('データベースのみリセット')
  .action(async () => {
    try {
      await resetDatabase();
    } catch (error) {
      console.error('エラー:', error);
    }
  });

resetCmd
  .command('documents')
  .description('ドキュメントファイルのみリセット')
  .action(async () => {
    try {
      await resetDocuments();
    } catch (error) {
      console.error('エラー:', error);
    }
  });

// AI commands
const aiCmd = program
  .command('ai')
  .description('AI機能');

aiCmd
  .command('generate-document <fragmentId>')
  .description('フラグメントからドキュメントを生成')
  .action(async (fragmentId) => {
    try {
      const documents = await generateDocumentFromFragment(parseInt(fragmentId));
      console.log('ドキュメントが生成されました:', documents);
    } catch (error) {
      console.error('エラー:', error);
    }
  });


program.parse();