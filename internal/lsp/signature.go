package lsp

import (
	"strings"
)

type SignatureProvider struct{}

func NewSignatureProvider() *SignatureProvider {
	return &SignatureProvider{}
}

type signatureEntry struct {
	label         string
	documentation string
	parameters    []ParameterInformation
}

func (sp *SignatureProvider) signatures() []signatureEntry {
	return []signatureEntry{
		{
			label:         "$if{condition: block}",
			documentation: "Conditional block evaluated against environment variables.\n\n```yaml\n$if{ IS_PRODUCTION }:\n  environment: production\n$endif\n```",
			parameters: []ParameterInformation{
				{Label: "condition: block", Documentation: "Condition expression followed by block content"},
			},
		},
		{
			label:         "$fn{upper(value: string)}",
			documentation: "Convert a string to uppercase.\n\n```yaml\nname: $fn{upper(project_name)}\n```",
			parameters: []ParameterInformation{
				{Label: "value: string", Documentation: "String value to convert to uppercase"},
			},
		},
		{
			label:         "$fn{lower(value: string)}",
			documentation: "Convert a string to lowercase.\n\n```yaml\nname: $fn{lower(project_name)}\n```",
			parameters: []ParameterInformation{
				{Label: "value: string", Documentation: "String value to convert to lowercase"},
			},
		},
		{
			label:         "$fn{slug(value: string)}",
			documentation: "Convert a string to a URL-safe slug.\n\n```yaml\npath: $fn{slug(project_name)}\n```",
			parameters: []ParameterInformation{
				{Label: "value: string", Documentation: "String value to slugify"},
			},
		},
		{
			label:         "$fn{default(value: any, fallback: any)}",
			documentation: "Return the first non-empty value, or the fallback.\n\n```yaml\nport: $fn{default($env{PORT}, 8080)}\n```",
			parameters: []ParameterInformation{
				{Label: "value: any", Documentation: "Primary value to check"},
				{Label: "fallback: any", Documentation: "Fallback value if primary is empty"},
			},
		},
	}
}

func (sp *SignatureProvider) Provide(text string, line, character int) *SignatureHelp {
	lines := strings.Split(text, "\n")
	if line < 0 || line >= len(lines) {
		return nil
	}

	lineText := lines[line]
	if character > len(lineText) {
		character = len(lineText)
	}
	before := lineText[:character]

	sigs := sp.signatures()
	var matching []SignatureInformation

	for _, s := range sigs {
		for _, trigger := range []string{"$if", "$fn{upper", "$fn{lower", "$fn{slug", "$fn{default"} {
			if strings.Contains(before, trigger) {
				info := SignatureInformation{
					Label:         s.label,
					Documentation: s.documentation,
					Parameters:    s.parameters,
				}
				if strings.HasSuffix(before, "(") || strings.HasSuffix(before, "{$") || strings.HasSuffix(before, "{") {
					if strings.Contains(trigger, "$if") && strings.Contains(before, "$if") {
						matching = append(matching, info)
					} else if strings.Contains(trigger, "$fn") && strings.Contains(before, "$fn") {
						matching = append(matching, info)
					}
				}
			}
		}
	}

	if len(matching) == 0 {
		for _, s := range sigs {
			info := SignatureInformation{
				Label:         s.label,
				Documentation: s.documentation,
				Parameters:    s.parameters,
			}
			matching = append(matching, info)
		}
		if len(matching) == 0 {
			return nil
		}
	}

	return &SignatureHelp{
		Signatures:      matching,
		ActiveSignature: 0,
		ActiveParameter: 0,
	}
}
