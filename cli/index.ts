#!/usr/bin/env node

import { Command } from 'commander';
import { FragmentService, DocumentService, TagService, QuestionService, AIService } from '../core/index.js';

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
    const fragmentService = new FragmentService();
    try {
      const fragment = await fragmentService.create({
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
    const fragmentService = new FragmentService();
    try {
      const fragments = await fragmentService.findAll();
      console.log('フラグメント一覧:');
      fragments.forEach(f => {
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
    const fragmentService = new FragmentService();
    try {
      const fragment = await fragmentService.findById(parseInt(id));
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
    const documentService = new DocumentService();
    try {
      const document = await documentService.create({
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
    const documentService = new DocumentService();
    try {
      const documents = await documentService.findAll();
      console.log('ドキュメント一覧:');
      documents.forEach(d => {
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
    const documentService = new DocumentService();
    try {
      const document = await documentService.findById(parseInt(id));
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
    const tagService = new TagService();
    try {
      const tag = await tagService.create({
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
    const tagService = new TagService();
    try {
      const tags = await tagService.findAll();
      console.log('タグ一覧:');
      tags.forEach(t => {
        console.log(`ID: ${t.id}, 名前: ${t.name}`);
      });
    } catch (error) {
      console.error('エラー:', error);
    }
  });

// Question commands
const questionCmd = program
  .command('question')
  .description('質問管理');

questionCmd
  .command('create')
  .description('新しい質問を作成')
  .option('-c, --content <content>', '質問の内容')
  .action(async (options) => {
    const questionService = new QuestionService();
    try {
      const question = await questionService.create({
        content: options.content,
      });
      console.log('質問が作成されました:', question);
    } catch (error) {
      console.error('エラー:', error);
    }
  });

questionCmd
  .command('list')
  .description('全ての質問を表示')
  .action(async () => {
    const questionService = new QuestionService();
    try {
      const questions = await questionService.findAll();
      console.log('質問一覧:');
      questions.forEach(q => {
        console.log(`ID: ${q.id}, 内容: ${q.content}`);
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
      const { migrate } = await import('drizzle-orm/libsql/migrator');
      const { db } = await import('../core/db/index.js');
      
      await migrate(db, { migrationsFolder: './drizzle' });
      console.log('データベースが初期化されました');
    } catch (error) {
      console.error('データベース初期化エラー:', error);
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
    const aiService = new AIService();
    try {
      const document = await aiService.generateDocumentFromFragment(parseInt(fragmentId));
      console.log('ドキュメントが生成されました:', document);
    } catch (error) {
      console.error('エラー:', error);
    }
  });

aiCmd
  .command('generate-tags <documentId>')
  .description('ドキュメントにタグを生成')
  .action(async (documentId) => {
    const aiService = new AIService();
    try {
      const tags = await aiService.generateTagsForDocument(parseInt(documentId));
      console.log('タグが生成されました:', tags);
    } catch (error) {
      console.error('エラー:', error);
    }
  });

aiCmd
  .command('summarize')
  .description('テキストを要約')
  .option('-t, --text <text>', '要約したいテキスト')
  .action(async (options) => {
    const aiService = new AIService();
    try {
      const summary = await aiService.summarizeContent(options.text);
      console.log('要約:', summary);
    } catch (error) {
      console.error('エラー:', error);
    }
  });

aiCmd
  .command('generate-questions <fragmentId>')
  .description('フラグメントから質問を生成')
  .action(async (fragmentId) => {
    const aiService = new AIService();
    try {
      const questions = await aiService.generateQuestionsFromFragment(parseInt(fragmentId));
      console.log('生成された質問:');
      questions.forEach((q, i) => {
        console.log(`${i + 1}. ${q}`);
      });
    } catch (error) {
      console.error('エラー:', error);
    }
  });

aiCmd
  .command('process-all')
  .description('未処理のフラグメントを全て処理')
  .action(async () => {
    const aiService = new AIService();
    try {
      const documents = await aiService.processUnprocessedFragments();
      console.log(`${documents.length}個のドキュメントを生成しました`);
      documents.forEach(doc => {
        console.log(`- ${doc.title}`);
      });
    } catch (error) {
      console.error('エラー:', error);
    }
  });

program.parse();