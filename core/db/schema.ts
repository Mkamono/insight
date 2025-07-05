import { text, integer, sqliteTable, primaryKey, index } from 'drizzle-orm/sqlite-core';
import { relations } from 'drizzle-orm';

export const fragments = sqliteTable('fragments', {
  id: integer('id').primaryKey({ autoIncrement: true }),
  content: text('content').notNull(),
  url: text('url'),
  imagePath: text('image_path'),
  processed: integer('processed', { mode: 'boolean' }).default(false),
  parentId: integer('parent_id'),
  createdAt: integer('created_at', { mode: 'timestamp' }).$defaultFn(() => new Date()),
  updatedAt: integer('updated_at', { mode: 'timestamp' }).$defaultFn(() => new Date()),
}, (table) => ({
  processedIdx: index('fragments_processed_idx').on(table.processed),
  parentIdIdx: index('fragments_parent_id_idx').on(table.parentId),
  createdAtIdx: index('fragments_created_at_idx').on(table.createdAt),
}));

export const documents = sqliteTable('documents', {
  id: integer('id').primaryKey({ autoIncrement: true }),
  title: text('title').notNull().unique(),
  content: text('content').notNull(),
  summary: text('summary').notNull(),
  createdAt: integer('created_at', { mode: 'timestamp' }).$defaultFn(() => new Date()),
  updatedAt: integer('updated_at', { mode: 'timestamp' }).$defaultFn(() => new Date()),
}, (table) => ({
  titleIdx: index('documents_title_idx').on(table.title),
  createdAtIdx: index('documents_created_at_idx').on(table.createdAt),
  updatedAtIdx: index('documents_updated_at_idx').on(table.updatedAt),
}));

export const tags = sqliteTable('tags', {
  id: integer('id').primaryKey({ autoIncrement: true }),
  name: text('name').notNull().unique(),
  createdAt: integer('created_at', { mode: 'timestamp' }).$defaultFn(() => new Date()),
  updatedAt: integer('updated_at', { mode: 'timestamp' }).$defaultFn(() => new Date()),
}, (table) => ({
  nameIdx: index('tags_name_idx').on(table.name),
}));

export const questions = sqliteTable('questions', {
  id: integer('id').primaryKey({ autoIncrement: true }),
  content: text('content').notNull(),
  createdAt: integer('created_at', { mode: 'timestamp' }).$defaultFn(() => new Date()),
  updatedAt: integer('updated_at', { mode: 'timestamp' }).$defaultFn(() => new Date()),
});

export const fragmentDocuments = sqliteTable('fragment_documents', {
  fragmentId: integer('fragment_id').notNull().references(() => fragments.id, { onDelete: 'cascade' }),
  documentId: integer('document_id').notNull().references(() => documents.id, { onDelete: 'cascade' }),
}, (table) => ({
  pk: primaryKey({ columns: [table.fragmentId, table.documentId] }),
}));

export const documentTags = sqliteTable('document_tags', {
  documentId: integer('document_id').notNull().references(() => documents.id, { onDelete: 'cascade' }),
  tagId: integer('tag_id').notNull().references(() => tags.id, { onDelete: 'cascade' }),
}, (table) => ({
  pk: primaryKey({ columns: [table.documentId, table.tagId] }),
}));

// Relations
export const fragmentsRelations = relations(fragments, ({ one, many }) => ({
  parent: one(fragments, {
    fields: [fragments.parentId],
    references: [fragments.id],
  }),
  children: many(fragments),
}));

export const documentsRelations = relations(documents, ({ many }) => ({
  fragmentDocuments: many(fragmentDocuments),
  documentTags: many(documentTags),
}));

export const tagsRelations = relations(tags, ({ many }) => ({
  documentTags: many(documentTags),
}));

export const fragmentDocumentsRelations = relations(fragmentDocuments, ({ one }) => ({
  fragment: one(fragments, {
    fields: [fragmentDocuments.fragmentId],
    references: [fragments.id],
  }),
  document: one(documents, {
    fields: [fragmentDocuments.documentId],
    references: [documents.id],
  }),
}));

export const documentTagsRelations = relations(documentTags, ({ one }) => ({
  document: one(documents, {
    fields: [documentTags.documentId],
    references: [documents.id],
  }),
  tag: one(tags, {
    fields: [documentTags.tagId],
    references: [tags.id],
  }),
}));

export type Fragment = typeof fragments.$inferSelect;
export type NewFragment = typeof fragments.$inferInsert;
export type Document = typeof documents.$inferSelect;
export type NewDocument = typeof documents.$inferInsert;
export type Tag = typeof tags.$inferSelect;
export type NewTag = typeof tags.$inferInsert;
export type Question = typeof questions.$inferSelect;
export type NewQuestion = typeof questions.$inferInsert;