## naeos search

Full-text search engine management

### Synopsis

Manage search indexes, query documents, and perform full-text search.

Example:
  naeos search index --id doc1 --title "Hello World" --content "This is a test"
  naeos search query --term "hello"
  naeos search count
  naeos search delete --id doc1
  naeos search list

```
naeos search [flags]
```

### Options

```
  -h, --help   help for search
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos](naeos.md)	 - NAEOS CLI - Declarative Engineering Runtime
* [naeos search count](naeos_search_count.md)	 - Count documents in index
* [naeos search delete](naeos_search_delete.md)	 - Delete a document from index
* [naeos search index](naeos_search_index.md)	 - Index a document
* [naeos search list](naeos_search_list.md)	 - List all search indexes
* [naeos search query](naeos_search_query.md)	 - Search for documents

