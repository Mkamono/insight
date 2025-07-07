import { generateObject, generateText } from 'ai';
import { z } from 'zod';
import type { Document, Fragment } from '../db/index.js';
import { getAiModel } from './aimodel.js';
import { findFragmentById, findFragmentHierarchy, findFragmentWithChildren } from './fragmentService.js';
import { findAllDocuments } from './documentService.js';
import { createDocumentTool, updateDocumentTool, getDocumentDetailTool, createQuestionTool } from './tools.js';

export async function generateDocumentFromFragment(fragmentIds: number | number[]): Promise<Document[]> {
  // 単一IDの場合は配列に変換
  const ids = Array.isArray(fragmentIds) ? fragmentIds : [fragmentIds];
  
  // フラグメントを取得（親子関係付き）
  const fragments: Fragment[] = [];
  const fragmentsWithContext: string[] = [];
  
  for (const id of ids) {
    const fragment = await findFragmentById(id);
    if (!fragment) {
      throw new Error(`Fragment with id ${id} not found`);
    }
    fragments.push(fragment);
    
    // 親子関係の文脈を構築
    if (fragment.parentId) {
      // 子フラグメントの場合、階層構造を取得
      const hierarchy = await findFragmentHierarchy(id);
      const contextStr = hierarchy.map((f, index) => {
        const role = index === 0 ? '[親投稿]' : index === hierarchy.length - 1 ? '[この投稿]' : '[返信]';
        return `${role} ID: ${f.id} - ${f.content}`;
      }).join('\n↓ 返信として\n');
      fragmentsWithContext.push(`会話スレッド:\n${contextStr}`);
    } else {
      // ルートフラグメントの場合、子も含めて表示
      const withChildren = await findFragmentWithChildren(id);
      if (withChildren && withChildren.children.length > 0) {
        let contextStr = `[元投稿] ID: ${fragment.id} - ${fragment.content}`;
        withChildren.children.forEach(child => {
          contextStr += `\n↓ 返信として\n[返信] ID: ${child.id} - ${child.content}`;
        });
        fragmentsWithContext.push(`会話スレッド:\n${contextStr}`);
      } else {
        fragmentsWithContext.push(`単独投稿:\nID: ${fragment.id} - ${fragment.content}`);
      }
    }
  }

  const existingDocuments = await findAllDocuments();

  // 既存ドキュメントの情報を簡潔にまとめる
  const existingDocsInfo = existingDocuments.map(doc => ({
    id: doc.id,
    title: doc.title,
    summary: doc.summary
  }));

  const model = getAiModel();
  const prompt = `
複数のフラグメントを効率的に処理して、適切にドキュメントを作成・更新してください。
フラグメントには親子関係（スレッド構造）があります。文脈を考慮してドキュメント化してください。

未処理フラグメント（親子関係付き）:
${fragmentsWithContext.join('\n\n')}

既存ドキュメント:
${existingDocsInfo.length > 0 ? existingDocsInfo.map(doc =>
    `ID: ${doc.id}, タイトル: ${doc.title}, 要約: ${doc.summary}`
  ).join('\n') : '既存ドキュメントはありません'}

## 重要な理解事項
- 「会話スレッド」は一連の投稿と返信の流れです
- [元投稿]は話題の発端で、[返信]はそれに対する応答です
- 「↓ 返信として」は投稿の時系列と関係性を示します
- スレッド全体で一つの議論や情報交換が行われています

## 処理指示:
1. **会話の文脈を理解**: スレッドは投稿→返信の時系列的な会話として処理してください
2. **スレッド全体を一つの単位**として考え、関連するドキュメントを判断してください
3. 会話の内容が既存ドキュメントと関連する場合は更新を、新しいトピックの場合は新規作成してください
4. **スレッド内の全フラグメントID**をfragmentIdsに含めてください（会話全体を文書化するため）
5. ドキュメントの内容は会話の流れを反映し、質問→回答、問題→解決策などの構造を明確にしてください
6. **情報が不足している場合**: 
   - フラグメントの内容が曖昧、不完全、または詳細が欠けている場合
   - ドキュメント化に必要な重要情報が明らかに不足している場合
   - createQuestionツールを使ってユーザーに具体的な質問を作成してください
7. 適切なタグも提案してください
8. **重要**: contentパラメータはMarkdown形式で記述してください（見出し、リスト、強調などを使用）

処理例:
- createDocument(title: "Python概要", content: "# Python\n\n## 特徴\n- 可読性に優れている\n- 動的型付け\n\n## 用途\nWebアプリケーション開発等", summary: "...", tags: [...], fragmentIds: [1, 2])

利用可能なツール:
- getDocumentDetail: 特定ドキュメントの詳細を取得
- createDocument: 新しいドキュメントを作成（content=Markdown形式、fragmentIds必須）
- updateDocument: 既存ドキュメントを更新（content=Markdown形式、fragmentIds必須）
- createQuestion: 情報が不足している場合にユーザーに質問を作成（詳細な情報収集のため）
`;

  await generateText({
    model,
    prompt,
    tools: {
      getDocumentDetail: getDocumentDetailTool,
      createDocument: createDocumentTool,
      updateDocument: updateDocumentTool,
      createQuestion: createQuestionTool,
    },
    maxSteps: 20,
  });

  // 処理されたドキュメントを取得
  const updatedDocuments = await findAllDocuments();
  const processedDocuments = updatedDocuments.filter(doc => {
    const existingDoc = existingDocuments.find(existing => existing.id === doc.id);
    if (!existingDoc) {
      // 新規ドキュメント
      return true;
    }
    // 更新されたドキュメント（時間比較）
    return existingDoc.updatedAt && doc.updatedAt &&
      existingDoc.updatedAt.getTime() !== doc.updatedAt.getTime();
  });

  return processedDocuments;
}




