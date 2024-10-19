package models

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	varRegex    = regexp.MustCompile(`<aic:var\s+name="([^"]+)"(?:\s+default="([^"]*)")?(?:\s+options="([^"]*)")?\s*/>`)
	promptRegex = regexp.MustCompile(`<aic:prompt\s+id="([^"@]+)(?:@(\d+))?"\s*([^>]*)/>`)
)

type PromptVariable struct {
	Name    string
	Default string
	Options []string
}

func ExtractVariables(content string) []PromptVariable {
	matches := varRegex.FindAllStringSubmatch(content, -1)
	variables := make([]PromptVariable, 0, len(matches))

	for _, match := range matches {
		variable := PromptVariable{
			Name:    match[1],
			Default: match[2],
		}
		if match[3] != "" {
			variable.Options = strings.Split(match[3], ",")
		}
		variables = append(variables, variable)
	}

	return variables
}

func ReplaceVariables(content string, values map[string]string) string {
	return varRegex.ReplaceAllStringFunc(content, func(match string) string {
		submatch := varRegex.FindStringSubmatch(match)
		name := submatch[1]
		defaultValue := submatch[2]

		if value, ok := values[name]; ok {
			return value
		}
		if defaultValue != "" {
			return defaultValue
		}
		return fmt.Sprintf("[Missing value for %s]", name)
	})
}

func ExtractPromptReferences(content string) map[string]map[string]string {
	matches := promptRegex.FindAllStringSubmatch(content, -1)
	references := make(map[string]map[string]string)

	for _, match := range matches {
		promptID := match[1]
		version := match[2]
		vars := parseKeyValuePairs(match[3])

		reference := make(map[string]string)
		if version != "" {
			reference["version"] = version
		}
		for k, v := range vars {
			reference[k] = v
		}

		references[promptID] = reference
	}

	return references
}

func parseKeyValuePairs(s string) map[string]string {
	pairs := make(map[string]string)
	parts := strings.Fields(s)
	for _, part := range parts {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) == 2 {
			key := strings.Trim(kv[0], `"`)
			value := strings.Trim(kv[1], `"`)
			pairs[key] = value
		}
	}
	return pairs
}

func ReplacePromptReferences(content string, prompts map[string]*AgentPrompt) (string, error) {
	return promptRegex.ReplaceAllStringFunc(content, func(match string) string {
		submatch := promptRegex.FindStringSubmatch(match)
		promptID := submatch[1]
		version := submatch[2]
		vars := parseKeyValuePairs(submatch[3])

		prompt, ok := prompts[promptID]
		if !ok {
			return fmt.Sprintf("[Prompt not found: %s]", promptID)
		}

		var promptVersion *PromptVersion
		if version == "" {
			promptVersion = prompt.GetVersion(0) // Latest version
		} else {
			versionInt := 0
			fmt.Sscanf(version, "%d", &versionInt)
			promptVersion = prompt.GetVersion(versionInt)
		}

		if promptVersion == nil {
			return fmt.Sprintf("[Version not found for prompt: %s@%s]", promptID, version)
		}

		replacedContent := ReplaceVariables(promptVersion.Content, vars)
		return replacedContent
	}), nil
}
