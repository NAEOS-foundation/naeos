package ai

import "testing"

func TestAI_ZeroValue(t *testing.T) {
	var a AI
	if a.Models != nil {
		t.Error("expected nil Models")
	}
	if a.Prompts != nil {
		t.Error("expected nil Prompts")
	}
}

func TestAI_Full(t *testing.T) {
	a := AI{
		Models: []Model{
			{Name: "gpt4", Kind: "llm", Version: "1.0"},
		},
		Prompts: []Prompt{
			{Name: "enrich", Template: "template", Kind: "llm"},
		},
		ContextBundles: []ContextBundle{
			{Name: "ctx1", Sources: []string{"src1"}},
		},
		Embeddings: []Embedding{
			{Name: "emb1", Dimension: 768},
		},
		Attributes: map[string]string{"key": "val"},
	}
	if len(a.Models) != 1 {
		t.Errorf("expected 1 model, got %d", len(a.Models))
	}
	if a.Models[0].Name != "gpt4" {
		t.Errorf("expected gpt4, got %s", a.Models[0].Name)
	}
	if len(a.Prompts) != 1 {
		t.Errorf("expected 1 prompt, got %d", len(a.Prompts))
	}
	if len(a.ContextBundles) != 1 {
		t.Errorf("expected 1 context bundle, got %d", len(a.ContextBundles))
	}
	if a.Embeddings[0].Dimension != 768 {
		t.Errorf("expected 768, got %d", a.Embeddings[0].Dimension)
	}
	if a.Attributes["key"] != "val" {
		t.Errorf("expected val, got %s", a.Attributes["key"])
	}
}
