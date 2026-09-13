// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package generation

import "github.com/NAEOS-foundation/naeos/internal/neir/model/language"

type GenerationConfig struct {
	Languages []language.Language `json:"languages,omitempty"`
	OutputDir string              `json:"output_dir,omitempty"`
	ModuleDir string              `json:"module_dir,omitempty"`
}

func (gc *GenerationConfig) HasLanguage(lang language.Language) bool {
	for _, l := range gc.Languages {
		if l == lang {
			return true
		}
	}
	return false
}

func (gc *GenerationConfig) DefaultLanguage() language.Language {
	if len(gc.Languages) > 0 {
		return gc.Languages[0]
	}
	return language.LanguageGo
}
