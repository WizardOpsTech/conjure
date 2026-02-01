package prompt

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/wizardopstech/conjure/internal/metadata"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("99"))

	promptStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86"))

	descriptionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("241"))

	inputStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("219"))

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42"))
)

type model struct {
	metadata     *metadata.TemplateMetadata
	currentIndex int
	values       map[string]interface{}
	currentInput string
	finished     bool
	err          error
}

func initialModel(meta *metadata.TemplateMetadata) model {
	return model{
		metadata:     meta,
		currentIndex: 0,
		values:       make(map[string]interface{}),
		currentInput: "",
		finished:     false,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.err = fmt.Errorf("cancelled by user")
			m.finished = true
			return m, tea.Quit

		case "enter":
			if m.currentIndex >= len(m.metadata.Variables) {
				return m, nil
			}

			currentVar := m.metadata.Variables[m.currentIndex]

			if m.currentInput == "" && currentVar.Default != "" {
				m.values[currentVar.Name] = currentVar.Default
			} else if m.currentInput == "" && currentVar.Default == "" {
				return m, nil
			} else if m.currentInput != "" {
				m.values[currentVar.Name] = m.currentInput
			}

			m.currentIndex++
			m.currentInput = ""

			if m.currentIndex >= len(m.metadata.Variables) {
				m.finished = true
				return m, tea.Quit
			}

		case "backspace":
			if len(m.currentInput) > 0 {
				m.currentInput = m.currentInput[:len(m.currentInput)-1]
			}

		default:
			if len(msg.String()) == 1 {
				m.currentInput += msg.String()
			}
		}
	}

	return m, nil
}

func (m model) View() string {
	if m.finished {
		if m.err != nil {
			return ""
		}
		return successStyle.Render("✓ All variables collected!\n")
	}

	var b strings.Builder

	// Title
	_, _ = b.WriteString(titleStyle.Render(fmt.Sprintf("Template: %s", m.metadata.TemplateName)))
	_, _ = b.WriteString("\n")
	_, _ = b.WriteString(descriptionStyle.Render(m.metadata.TemplateDescription))
	_, _ = b.WriteString("\n\n")

	// Progress
	_, _ = b.WriteString(descriptionStyle.Render(fmt.Sprintf("Variable %d of %d", m.currentIndex+1, len(m.metadata.Variables))))
	_, _ = b.WriteString("\n\n")

	// Current variable
	if m.currentIndex < len(m.metadata.Variables) {
		currentVar := m.metadata.Variables[m.currentIndex]

		// Variable name
		required := ""
		if currentVar.Default == "" {
			required = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(" *")
		}
		_, _ = b.WriteString(promptStyle.Render(fmt.Sprintf("%s%s", currentVar.Name, required)))
		_, _ = b.WriteString("\n")

		// Description
		_, _ = b.WriteString(descriptionStyle.Render(currentVar.Description))
		_, _ = b.WriteString("\n")

		// Default value hint
		if currentVar.Default != "" {
			_, _ = b.WriteString(descriptionStyle.Render(fmt.Sprintf("[default: %s]", currentVar.Default)))
			_, _ = b.WriteString("\n")
		}

		// Input field
		_, _ = b.WriteString(inputStyle.Render("> " + m.currentInput + "█"))
		_, _ = b.WriteString("\n\n")

		// Help text
		_, _ = b.WriteString(descriptionStyle.Render("Press Enter to continue, Ctrl+C to cancel"))
	}

	return b.String()
}

func CollectVariables(meta *metadata.TemplateMetadata, existingVars map[string]interface{}) (map[string]interface{}, error) {
	m := initialModel(meta)

	if existingVars != nil {
		m.values = existingVars
	}

	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("error running prompt: %w", err)
	}

	final := finalModel.(model)
	if final.err != nil {
		return nil, final.err
	}

	return final.values, nil
}

func CollectBundleVariables(bundleMeta *metadata.BundleMetadata, existingVars map[string]interface{}) (map[string]interface{}, map[string]map[string]interface{}, error) {
	varToTemplates := make(map[string][]string)

	for _, v := range bundleMeta.SharedVariables {
		varToTemplates[v.Name] = []string{"(shared)"}
	}

	for templateName, templateVars := range bundleMeta.TemplateVariables {
		for _, v := range templateVars {
			if templates, exists := varToTemplates[v.Name]; exists {
				varToTemplates[v.Name] = append(templates, templateName)
			} else {
				varToTemplates[v.Name] = []string{templateName}
			}
		}
	}

	allVars := metadata.GetAllVariablesForBundle(bundleMeta)
	for i := range allVars {
		if templates, exists := varToTemplates[allVars[i].Name]; exists {
			templateList := strings.Join(templates, ", ")
			if allVars[i].Description != "" {
				allVars[i].Description = fmt.Sprintf("%s\n  Used in: %s", allVars[i].Description, templateList)
			} else {
				allVars[i].Description = fmt.Sprintf("Used in: %s", templateList)
			}
		}
	}

	tempMeta := &metadata.TemplateMetadata{
		TemplateName:        bundleMeta.BundleName,
		TemplateDescription: bundleMeta.BundleDescription,
		TemplateType:        bundleMeta.BundleType,
		Variables:           allVars,
	}

	vars, err := CollectVariables(tempMeta, existingVars)
	if err != nil {
		return nil, nil, err
	}

	overrides, err := CollectTemplateOverrides(bundleMeta, vars)
	if err != nil {
		return nil, nil, err
	}

	return vars, overrides, nil
}

func CollectTemplateOverrides(bundleMeta *metadata.BundleMetadata, currentVars map[string]interface{}) (map[string]map[string]interface{}, error) {
	overrides := make(map[string]map[string]interface{})

	templateNames := make([]string, 0)
	for templateName := range bundleMeta.TemplateVariables {
		templateNames = append(templateNames, templateName)
	}

	sharedVarNames := make([]string, 0)
	for _, v := range bundleMeta.SharedVariables {
		sharedVarNames = append(sharedVarNames, v.Name)
	}

	if len(sharedVarNames) == 0 {
		return overrides, nil
	}

	fmt.Println()
	fmt.Println(descriptionStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(titleStyle.Render("Template-Specific Overrides"))
	fmt.Println(descriptionStyle.Render("Override shared variables for specific templates"))
	fmt.Println()

	for {
		fmt.Print(promptStyle.Render("Add a template-specific override? (y/n): "))
		var response string
		if _, err := fmt.Scanln(&response); err != nil {
			break
		}
		response = strings.TrimSpace(strings.ToLower(response))

		if response != "y" && response != "yes" {
			break
		}

		fmt.Println()
		fmt.Println(descriptionStyle.Render("Available templates:"))
		for i, tmpl := range templateNames {
			fmt.Printf("  %d. %s\n", i+1, tmpl)
		}
		fmt.Print(promptStyle.Render("Select template (number): "))
		var templateIdx int
		if _, err := fmt.Scanln(&templateIdx); err != nil {
			fmt.Println(descriptionStyle.Render("Invalid input, skipping..."))
			continue
		}
		if templateIdx < 1 || templateIdx > len(templateNames) {
			fmt.Println(descriptionStyle.Render("Invalid selection, skipping..."))
			continue
		}
		selectedTemplate := templateNames[templateIdx-1]

		fmt.Println()
		fmt.Println(descriptionStyle.Render("Shared variables:"))
		for i, varName := range sharedVarNames {
			currentVal := currentVars[varName]
			fmt.Printf("  %d. %s (current: %v)\n", i+1, varName, currentVal)
		}
		fmt.Print(promptStyle.Render("Select variable to override (number): "))
		var varIdx int
		if _, err := fmt.Scanln(&varIdx); err != nil {
			fmt.Println(descriptionStyle.Render("Invalid input, skipping..."))
			continue
		}
		if varIdx < 1 || varIdx > len(sharedVarNames) {
			fmt.Println(descriptionStyle.Render("Invalid selection, skipping..."))
			continue
		}
		selectedVar := sharedVarNames[varIdx-1]

		fmt.Print(promptStyle.Render(fmt.Sprintf("New value for %s in %s: ", selectedVar, selectedTemplate)))
		var newValue string
		if _, err := fmt.Scanln(&newValue); err != nil {
			fmt.Println(descriptionStyle.Render("Invalid input, skipping..."))
			continue
		}
		newValue = strings.TrimSpace(newValue)

		if newValue == "" {
			fmt.Println(descriptionStyle.Render("Empty value, skipping..."))
			continue
		}

		if overrides[selectedTemplate] == nil {
			overrides[selectedTemplate] = make(map[string]interface{})
		}
		overrides[selectedTemplate][selectedVar] = newValue

		fmt.Println(successStyle.Render(fmt.Sprintf("✓ Override added: %s.%s = %s", selectedTemplate, selectedVar, newValue)))
		fmt.Println()
	}

	if len(overrides) > 0 {
		fmt.Println()
		fmt.Println(successStyle.Render(fmt.Sprintf("✓ %d template override(s) configured", len(overrides))))
	}

	return overrides, nil
}
