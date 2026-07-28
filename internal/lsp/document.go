package lsp

import "sync"

type Document struct {
	URI     string
	Version int
	Text    string
}

type DocumentManager struct {
	mu        sync.RWMutex
	documents map[string]*Document
}

func NewDocumentManager() *DocumentManager {
	return &DocumentManager{
		documents: make(map[string]*Document),
	}
}

func (dm *DocumentManager) Open(uri string, version int, text string) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	dm.documents[uri] = &Document{
		URI:     uri,
		Version: version,
		Text:    text,
	}
}

func (dm *DocumentManager) Update(uri string, version int, text string) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	if doc, ok := dm.documents[uri]; ok {
		doc.Version = version
		doc.Text = text
	}
}

func (dm *DocumentManager) Close(uri string) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	delete(dm.documents, uri)
}

func (dm *DocumentManager) Get(uri string) (*Document, bool) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	doc, ok := dm.documents[uri]
	return doc, ok
}

func (dm *DocumentManager) All() []*Document {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	result := make([]*Document, 0, len(dm.documents))
	for _, doc := range dm.documents {
		result = append(result, doc)
	}
	return result
}
