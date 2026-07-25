package testgen

import (
	"fmt"

	"golang.org/x/text/cases"
	xlanguage "golang.org/x/text/language"

	"github.com/NAEOS-foundation/naeos/internal/generation/engine"
	naeoserr "github.com/NAEOS-foundation/naeos/internal/errors"
	"github.com/NAEOS-foundation/naeos/internal/neir/model"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/language"
	"github.com/NAEOS-foundation/naeos/internal/shared/strutil"
)

func GenerateTests(neir *model.NEIR) ([]engine.Artifact, error) {
	if neir == nil {
		return nil, naeoserr.New(naeoserr.ErrPipeline, "neir is nil")
	}

	var artifacts []engine.Artifact
	projectName := ""
	if neir.Project != nil {
		projectName = neir.Project.Name
	}

	for _, lang := range neir.Generation.Languages {
		exts := language.Extensions(lang)
		testExt := ".go"
		if len(exts) > 0 {
			testExt = exts[0]
		}

		for _, mod := range neir.Modules {
			pkg := strutil.Slugify(mod.Name)
			artifacts = append(artifacts, engine.Artifact{
				Path:    fmt.Sprintf("%s/%s_test%s", mod.Path, pkg, testExt),
				Content: []byte(generateTest(lang, mod.Name, projectName)),
			})
		}

		for _, svc := range neir.Services {
			svcDir := fmt.Sprintf("internal/%s", strutil.Slugify(svc.Name))
			pkg := strutil.Slugify(svc.Name)
			artifacts = append(artifacts, engine.Artifact{
				Path:    fmt.Sprintf("%s/%s_test%s", svcDir, pkg, testExt),
				Content: []byte(generateServiceTest(lang, svc.Name, projectName)),
			})
		}

		if len(neir.Modules) > 0 || len(neir.Services) > 0 {
			artifacts = append(artifacts, engine.Artifact{
				Path:    fmt.Sprintf("integration_test%s", testExt),
				Content: []byte(generateIntegrationTest(lang, projectName)),
			})
		}
	}

	return artifacts, nil
}

func generateTest(lang language.Language, moduleName, projectName string) string {
	pkg := strutil.Slugify(moduleName)
	switch lang {
	case language.LanguageGo:
		return fmt.Sprintf(`package %s

import "testing"

func Test%sModule(t *testing.T) {
	t.Run("initialization", func(t *testing.T) {
		t.Log("test for %s module in %s")
	})
}

func Test%sHandler(t *testing.T) {
	t.Run("handles request", func(t *testing.T) {
		t.Log("handler test for %s")
	})
}
`, pkg, cases.Title(xlanguage.English).String(moduleName), moduleName, projectName, cases.Title(xlanguage.English).String(moduleName), moduleName)
	case language.LanguageTypeScript:
		return fmt.Sprintf(`import { describe, it, expect } from 'vitest';

describe('%s module', () => {
  it('should initialize', () => {
    expect(true).toBe(true);
  });
});
`, moduleName)
	case language.LanguagePython:
		return fmt.Sprintf(`import unittest

class Test%s(unittest.TestCase):
    def test_initialization(self):
        self.assertTrue(True)

    def test_handler(self):
        self.assertTrue(True)

if __name__ == '__main__':
    unittest.main()
`, cases.Title(xlanguage.English).String(moduleName))
	case language.LanguageJava:
		return fmt.Sprintf(`import org.junit.jupiter.api.Test;
import static org.junit.jupiter.api.Assertions.*;

class %sTest {
    @Test
    void shouldInitialize() {
        assertTrue(true);
    }
}
`, cases.Title(xlanguage.English).String(moduleName))
	case language.LanguageRust:
		return fmt.Sprintf(`#[cfg(test)]
mod tests {
    #[test]
    fn test_%s() {
        assert!(true);
    }
}
`, pkg)
	default:
		return ""
	}
}

func generateServiceTest(lang language.Language, serviceName, projectName string) string {
	pkg := strutil.Slugify(serviceName)
	switch lang {
	case language.LanguageGo:
		return fmt.Sprintf(`package %s

import "testing"

func TestRun(t *testing.T) {
	t.Run("starts server", func(t *testing.T) {
		t.Log("server test for %s in %s")
	})
}
`, pkg, serviceName, projectName)
	case language.LanguageTypeScript:
		return fmt.Sprintf(`import { describe, it, expect } from 'vitest';

describe('%s service', () => {
  it('should start', () => {
    expect(true).toBe(true);
  });
});
`, serviceName)
	case language.LanguagePython:
		return fmt.Sprintf(`import unittest

class Test%sService(unittest.TestCase):
    def test_start(self):
        self.assertTrue(True)

if __name__ == '__main__':
    unittest.main()
`, cases.Title(xlanguage.English).String(serviceName))
	case language.LanguageJava:
		return fmt.Sprintf(`import org.junit.jupiter.api.Test;
import static org.junit.jupiter.api.Assertions.*;

class %sServiceTest {
    @Test
    void shouldStart() {
        assertTrue(true);
    }
}
`, cases.Title(xlanguage.English).String(serviceName))
	case language.LanguageRust:
		return fmt.Sprintf(`#[cfg(test)]
mod tests {
    #[test]
    fn test_%s_service() {
        assert!(true);
    }
}
`, pkg)
	default:
		return ""
	}
}

func generateIntegrationTest(lang language.Language, projectName string) string {
	switch lang {
	case language.LanguageGo:
		return fmt.Sprintf(`package main

import "testing"

func TestIntegration(t *testing.T) {
	t.Run("full pipeline", func(t *testing.T) {
		t.Log("integration test for %s")
	})
}
`, projectName)
	case language.LanguageTypeScript:
		return `import { describe, it, expect } from 'vitest';

describe('integration', () => {
  it('should work end to end', () => {
    expect(true).toBe(true);
  });
});
`
	case language.LanguagePython:
		return `import unittest

class TestIntegration(unittest.TestCase):
    def test_full_pipeline(self):
        self.assertTrue(True)

if __name__ == '__main__':
    unittest.main()
`
	case language.LanguageJava:
		return `import org.junit.jupiter.api.Test;
import static org.junit.jupiter.api.Assertions.*;

class IntegrationTest {
    @Test
    void shouldWorkEndToEnd() {
        assertTrue(true);
    }
}
`
	case language.LanguageRust:
		return `#[cfg(test)]
mod tests {
    #[test]
    fn test_integration() {
        assert!(true);
    }
}
`
	default:
		return ""
	}
}
