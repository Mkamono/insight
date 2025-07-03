# Insight CLI Command Reference

This document outlines the command-line interface for the `insight` application.

## 1. Input (`insight add`)

This is the primary command for adding new pieces of information (`fragments`) to the database. It's designed to be flexible, accepting various combinations of text, URLs, and images.

### Usage
```bash
insight add [text] [--url <url>] [--image <path>]
```

### Examples

#### Text only
```bash
insight add "A simple text fragment."
```
- **Stored JSON:** `{"text": "A simple text fragment."}`

#### URL with a comment
```bash
insight add "Interesting article on system design." --url "https://example.com/article"
```
- **Stored JSON:** `{"text": "Interesting article on system design.", "url": "https://example.com/article"}`

#### URL only
```bash
insight add --url "https://example.com/article"
```
- **Stored JSON:** `{"url": "https://example.com/article"}`

#### Image with a comment
```bash
insight add "Architecture diagram for the new feature." --image "~/images/diagram.png"
```
- **Stored JSON:** `{"text": "Architecture diagram for the new feature.", "image_path": "~/images/diagram.png"}`

#### All combined
```bash
insight add "This figure in the article is key." --url "https://example.com/article" --image "path/to/figure.png"
```
- **Stored JSON:** `{"text": "This figure in the article is key.", "url": "https://example.com/article", "image_path": "path/to/figure.png"}`


## 2. Processing (`insight process`)

This command triggers the core logic of the application. It takes unprocessed `fragments` and uses an AI agent to organize them into `documents`.

### Usage
```bash
insight process
```

### Options
- `--dry-run`: Preview the changes the AI intends to make without modifying the database.
- `--doc <id_or_title>`: Force an update on a specific document using any relevant unprocessed fragments.


## 3. Browsing & Searching

Commands for viewing and finding your organized knowledge.

### `insight doc list`
- **Description:** Lists all documents with their ID, title, and last update time.

### `insight doc show <id_or_title>`
- **Description:** Displays the full Markdown content of a specific document, along with its metadata (tags, source fragments).

### `insight tag list`
- **Description:** Shows all existing tags and the number of documents associated with each.

### `insight find --tag <tag_name>`
- **Description:** Finds and lists all documents tagged with a specific tag.

### `insight find --query "<search_term>"`
- **Description:** Performs a full-text search across all document titles and content.


## 4. Management

Commands for checking the status and managing the database.

### `insight status`
- **Description:** Shows a summary of the database, including counts of documents, fragments, and unprocessed fragments.

### `insight fragment list`
- **Description:** Lists all recorded fragments.
- **Option:** `--unprocessed`: Shows only the fragments that have not yet been processed into a document.
