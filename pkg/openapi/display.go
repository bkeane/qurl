package openapi

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#5B47E0")).
			Padding(0, 2)

	methodStyles = map[string]lipgloss.Style{
		"GET": lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#61AFEF")).
			Padding(0, 1),
		"POST": lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#98C379")).
			Padding(0, 1),
		"PUT": lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#E5C07B")).
			Padding(0, 1),
		"DELETE": lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#E06C75")).
			Padding(0, 1),
		"PATCH": lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#C678DD")).
			Padding(0, 1),
		"HEAD": lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#56B6C2")).
			Padding(0, 1),
		"OPTIONS": lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#ABB2BF")).
			Padding(0, 1),
	}

	pathStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E5C07B")).
			Bold(true)

	summaryStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ABB2BF"))

	sectionStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#61AFEF")).
			MarginTop(1)

	paramStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#98C379"))

	requiredStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E06C75")).
			Bold(true)

	descriptionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#ABB2BF")).
				MarginLeft(2)

	codeStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#2C323C")).
			Foreground(lipgloss.Color("#ABB2BF")).
			Padding(0, 1)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#5B47E0")).
			Padding(1).
			MarginTop(1).
			MarginBottom(1)
)

type Displayer struct {
	parser *Parser
}

func NewDisplayer(parser *Parser) *Displayer {
	return &Displayer{parser: parser}
}

func (d *Displayer) RenderIndex(paths []PathInfo) string {
	if len(paths) == 0 {
		return summaryStyle.Render("No paths found matching the filters")
	}

	info, _ := d.parser.GetInfo()
	var output strings.Builder

	if info != nil {
		title := fmt.Sprintf(" %s ", info.Title)
		if info.Version != "" {
			title += fmt.Sprintf("v%s ", info.Version)
		}
		output.WriteString(titleStyle.Render(title))
		output.WriteString("\n\n")

		// For index view, show a brief description (first sentence or first 100 chars)
		if info.Description != "" {
			description := info.Description
			// Take first sentence or first 100 characters, whichever is shorter
			if idx := strings.Index(description, ". "); idx > 0 && idx < 100 {
				description = description[:idx+1]
			} else if len(description) > 100 {
				description = description[:97] + "..."
			}
			output.WriteString(summaryStyle.Render(description))
			output.WriteString("\n\n")
		}
	}

	// Show authentication information if available
	authInfo := d.renderAuthInfo()
	if authInfo != "" {
		output.WriteString(authInfo)
		output.WriteString("\n")
	}

	output.WriteString(sectionStyle.Render("Endpoints"))
	output.WriteString("\n\n")

	currentPath := ""
	for _, path := range paths {
		if path.Path != currentPath {
			if currentPath != "" {
				output.WriteString("\n")
			}
			output.WriteString(pathStyle.Render(path.Path))
			output.WriteString("\n")
			currentPath = path.Path
		}

		methodStyle := getMethodStyle(path.Method)
		output.WriteString("  ")
		output.WriteString(methodStyle.Render(path.Method))

		if path.Summary != "" {
			output.WriteString("  ")
			output.WriteString(summaryStyle.Render(path.Summary))
		}
		output.WriteString("\n")
	}

	return output.String()
}

func (d *Displayer) RenderOperation(path PathInfo) string {
	var output strings.Builder

	methodStyle := getMethodStyle(path.Method)
	header := fmt.Sprintf("%s %s", methodStyle.Render(path.Method), pathStyle.Render(path.Path))
	output.WriteString(header)
	output.WriteString("\n\n")

	if path.Summary != "" {
		output.WriteString(summaryStyle.Render(path.Summary))
		output.WriteString("\n")
	}

	if path.Description != "" && path.Description != path.Summary {
		output.WriteString("\n")
		output.WriteString(descriptionStyle.Render(path.Description))
		output.WriteString("\n")
	}

	if len(path.Parameters) > 0 {
		output.WriteString("\n")
		output.WriteString(sectionStyle.Render("Parameters"))
		output.WriteString("\n\n")
		output.WriteString(d.renderParameters(path.Parameters))
	}

	if path.RequestBody != nil {
		output.WriteString("\n")
		output.WriteString(sectionStyle.Render("Request Body"))
		output.WriteString("\n\n")
		output.WriteString(d.renderRequestBody(path.RequestBody))
	}

	if path.Responses != nil {
		output.WriteString("\n")
		output.WriteString(sectionStyle.Render("Responses"))
		output.WriteString("\n\n")
		output.WriteString(d.renderResponses(path.Responses))
	}

	return boxStyle.Render(output.String())
}

func (d *Displayer) renderParameters(params []*v3.Parameter) string {
	var output strings.Builder

	paramsByLocation := make(map[string][]*v3.Parameter)
	for _, param := range params {
		if param.In != "" {
			paramsByLocation[param.In] = append(paramsByLocation[param.In], param)
		}
	}

	for _, location := range []string{"path", "query", "header", "cookie"} {
		if params, ok := paramsByLocation[location]; ok && len(params) > 0 {
			locationTitle := strings.ToUpper(location[:1]) + location[1:]
			output.WriteString(fmt.Sprintf("%s parameters:\n", locationTitle))
			for _, param := range params {
				output.WriteString(d.renderParameter(param))
			}
			output.WriteString("\n")
		}
	}

	return output.String()
}

func (d *Displayer) renderParameter(param *v3.Parameter) string {
	var output strings.Builder

	output.WriteString("  • ")
	if param.Name != "" {
		output.WriteString(paramStyle.Render(param.Name))
	}

	if param.Required != nil && *param.Required {
		output.WriteString(" ")
		output.WriteString(requiredStyle.Render("*required"))
	}

	if param.Schema != nil && param.Schema.Schema() != nil {
		schema := param.Schema.Schema()
		if len(schema.Type) > 0 {
			output.WriteString(" ")
			output.WriteString(codeStyle.Render(schema.Type[0]))
		}
		if schema.Format != "" {
			output.WriteString(" ")
			output.WriteString(codeStyle.Render(fmt.Sprintf("(%s)", schema.Format)))
		}
	}

	output.WriteString("\n")

	if param.Description != "" {
		output.WriteString(descriptionStyle.Render(fmt.Sprintf("    %s", param.Description)))
		output.WriteString("\n")
	}

	return output.String()
}

func (d *Displayer) renderRequestBody(body *v3.RequestBody) string {
	var output strings.Builder

	if body.Required != nil && *body.Required {
		output.WriteString(requiredStyle.Render("Required"))
		output.WriteString("\n\n")
	}

	if body.Description != "" {
		output.WriteString(descriptionStyle.Render(body.Description))
		output.WriteString("\n\n")
	}

	if body.Content != nil {
		for contentType, mediaType := range body.Content.FromOldest() {
			output.WriteString("Content Type: ")
			output.WriteString(codeStyle.Render(contentType))
			output.WriteString("\n")

			if mediaType.Schema != nil {
				output.WriteString("\n")
				output.WriteString(d.renderSchema(mediaType.Schema.Schema(), 0))
				output.WriteString("\n")
			}
			break // Show first content type's schema as example
		}
	}

	return output.String()
}

func (d *Displayer) renderResponses(responses *v3.Responses) string {
	var output strings.Builder

	if responses.Codes != nil {
		codes := responses.Codes
		for code, response := range codes.FromOldest() {
			output.WriteString(d.renderResponse(code, response))
		}
	}

	if responses.Default != nil {
		output.WriteString(d.renderResponse("default", responses.Default))
	}

	return output.String()
}

func (d *Displayer) renderResponse(code string, response *v3.Response) string {
	var output strings.Builder

	statusStyle := getStatusStyle(code)
	output.WriteString("  ")
	output.WriteString(statusStyle.Render(code))

	if response.Description != "" {
		output.WriteString(" - ")
		output.WriteString(summaryStyle.Render(response.Description))
	}

	output.WriteString("\n")

	if response.Content != nil {
		for contentType, mediaType := range response.Content.FromOldest() {
			output.WriteString("    Content Type: ")
			output.WriteString(codeStyle.Render(contentType))
			output.WriteString("\n")

			if mediaType.Schema != nil && code == "200" {
				schemaOutput := d.renderSchema(mediaType.Schema.Schema(), 2)
				if schemaOutput != "" {
					output.WriteString(schemaOutput)
					output.WriteString("\n")
				}
			}
			break // Show first content type
		}
	}

	return output.String()
}

func getMethodStyle(method string) lipgloss.Style {
	if style, ok := methodStyles[method]; ok {
		return style
	}
	return methodStyles["GET"]
}

func getStatusStyle(code string) lipgloss.Style {
	if code == "default" {
		return codeStyle
	}

	if len(code) > 0 {
		switch code[0] {
		case '2':
			return lipgloss.NewStyle().
				Foreground(lipgloss.Color("#98C379")).
				Bold(true)
		case '3':
			return lipgloss.NewStyle().
				Foreground(lipgloss.Color("#61AFEF")).
				Bold(true)
		case '4':
			return lipgloss.NewStyle().
				Foreground(lipgloss.Color("#E5C07B")).
				Bold(true)
		case '5':
			return lipgloss.NewStyle().
				Foreground(lipgloss.Color("#E06C75")).
				Bold(true)
		}
	}

	return codeStyle
}

func (d *Displayer) renderAuthInfo() string {
	var output strings.Builder

	schemes, err := d.parser.GetSecuritySchemes()
	if err != nil || schemes == nil {
		return ""
	}

	security := d.parser.GetSecurity()
	if len(security) == 0 {
		return ""
	}

	output.WriteString(sectionStyle.Render("Authentication"))
	output.WriteString("\n\n")

	// Show the security requirements
	for i, secReq := range security {
		if i > 0 {
			output.WriteString(" OR ")
		}

		if secReq == nil || secReq.Requirements == nil {
			output.WriteString(summaryStyle.Render("No authentication required"))
			continue
		}

		var reqNames []string
		for schemeName := range secReq.Requirements.FromOldest() {
			reqNames = append(reqNames, schemeName)
		}

		if len(reqNames) == 0 {
			output.WriteString(summaryStyle.Render("No authentication required"))
			continue
		}

		for j, schemeName := range reqNames {
			if j > 0 {
				output.WriteString(" + ")
			}

			if scheme := schemes.GetOrZero(schemeName); scheme != nil {
				schemeType := scheme.Type
				if schemeType != "" {
					switch schemeType {
					case "http":
						if scheme.Scheme != "" {
							output.WriteString(paramStyle.Render(fmt.Sprintf("%s (%s)", schemeName, scheme.Scheme)))
						} else {
							output.WriteString(paramStyle.Render(fmt.Sprintf("%s (HTTP)", schemeName)))
						}
					case "apiKey":
						location := "header"
						if scheme.In != "" {
							location = scheme.In
						}
						output.WriteString(paramStyle.Render(fmt.Sprintf("%s (API Key in %s)", schemeName, location)))
					case "oauth2":
						output.WriteString(paramStyle.Render(fmt.Sprintf("%s (OAuth2)", schemeName)))
					case "openIdConnect":
						output.WriteString(paramStyle.Render(fmt.Sprintf("%s (OpenID Connect)", schemeName)))
					default:
						output.WriteString(paramStyle.Render(fmt.Sprintf("%s (%s)", schemeName, schemeType)))
					}
				} else {
					output.WriteString(paramStyle.Render(schemeName))
				}
			} else {
				output.WriteString(paramStyle.Render(schemeName))
			}
		}
	}

	return output.String()
}

func (d *Displayer) renderSchema(schema *base.Schema, indent int) string {
	var output strings.Builder
	indentStr := strings.Repeat("  ", indent)

	if len(schema.Type) > 0 {
		schemaType := schema.Type[0]

		if schemaType == "object" {
			output.WriteString(indentStr)
			output.WriteString(paramStyle.Render("Example (JSON):"))
			output.WriteString("\n")
			output.WriteString(d.generateJSONExample(schema, indent+1))
		} else if schemaType == "array" {
			output.WriteString(indentStr)
			output.WriteString(paramStyle.Render("Array of:"))
			output.WriteString("\n")
			if schema.Items != nil && schema.Items.IsA() {
				output.WriteString(d.renderSchema(schema.Items.A.Schema(), indent+1))
			}
		} else {
			output.WriteString(indentStr)
			output.WriteString("Type: ")
			output.WriteString(codeStyle.Render(schemaType))
			if schema.Format != "" {
				output.WriteString(" ")
				output.WriteString(codeStyle.Render(fmt.Sprintf("(%s)", schema.Format)))
			}
			output.WriteString("\n")
		}
	}

	if schema.Description != "" {
		output.WriteString(indentStr)
		output.WriteString(descriptionStyle.Render(schema.Description))
		output.WriteString("\n")
	}

	return output.String()
}

func (d *Displayer) generateJSONExample(schema *base.Schema, indent int) string {
	var output strings.Builder
	indentStr := strings.Repeat("  ", indent)

	output.WriteString(indentStr)
	output.WriteString(codeStyle.Render("{"))
	output.WriteString("\n")

	if schema.Properties != nil {
		// Count properties first
		totalProps := 0
		for range schema.Properties.FromOldest() {
			totalProps++
		}

		i := 0
		for propName, propSchema := range schema.Properties.FromOldest() {
			output.WriteString(indentStr + "  ")
			output.WriteString(codeStyle.Render(fmt.Sprintf(`"%s": `, propName)))

			propValue := d.getExampleValue(propSchema.Schema())
			output.WriteString(codeStyle.Render(propValue))

			if i < totalProps-1 {
				output.WriteString(",")
			}

			isRequired := false
			for _, req := range schema.Required {
				if req == propName {
					isRequired = true
					break
				}
			}

			if isRequired {
				output.WriteString(" ")
				output.WriteString(requiredStyle.Render("// required"))
			}

			if propSchema.Schema().Description != "" {
				output.WriteString(" ")
				output.WriteString(summaryStyle.Render(fmt.Sprintf("// %s", propSchema.Schema().Description)))
			}

			output.WriteString("\n")
			i++
		}
	}

	output.WriteString(indentStr)
	output.WriteString(codeStyle.Render("}"))

	return output.String()
}

func (d *Displayer) getExampleValue(schema *base.Schema) string {
	if schema.Example != nil {
		// Try to extract a clean example value
		exampleStr := fmt.Sprintf("%v", schema.Example)
		// Clean up YAML node representations
		if strings.Contains(exampleStr, "!!") {
			// Extract the actual value from YAML node representation
			parts := strings.Fields(exampleStr)
			for i, part := range parts {
				if !strings.Contains(part, "!!") && !strings.Contains(part, "&{") && part != "<nil>" && part != "[]" {
					// Found a potential value
					if i > 0 && strings.Contains(parts[i-1], "!!str") {
						return fmt.Sprintf(`"%s"`, part)
					} else if i > 0 && (strings.Contains(parts[i-1], "!!int") || strings.Contains(parts[i-1], "!!float")) {
						return part
					}
				}
			}
		}
	}

	if len(schema.Type) > 0 {
		switch schema.Type[0] {
		case "string":
			if schema.Format == "date-time" {
				return `"2024-01-01T00:00:00Z"`
			} else if schema.Format == "date" {
				return `"2024-01-01"`
			} else if schema.Format == "email" {
				return `"user@example.com"`
			} else if schema.Format == "uri" || schema.Format == "url" {
				return `"https://example.com"`
			}
			if schema.Enum != nil && len(schema.Enum) > 0 {
				// Clean up enum value
				enumStr := fmt.Sprintf("%v", schema.Enum[0])
				if strings.Contains(enumStr, "!!str") {
					// Extract clean string from YAML node
					if idx := strings.Index(enumStr, "!!str"); idx >= 0 {
						remaining := enumStr[idx+5:]
						parts := strings.Fields(remaining)
						if len(parts) > 0 {
							clean := strings.Trim(parts[0], "{}")
							return fmt.Sprintf(`"%s"`, clean)
						}
					}
				}
				// Fallback: try to clean the value
				clean := strings.Trim(enumStr, "&{} ")
				if clean != "" && !strings.Contains(clean, "!!") {
					return fmt.Sprintf(`"%s"`, clean)
				}
				return `"string"`
			}
			return `"string"`
		case "number":
			if schema.Format == "float" {
				return "1.5"
			}
			return "123.45"
		case "integer":
			if schema.Format == "int64" {
				return "12345"
			} else if schema.Format == "int32" {
				return "123"
			}
			return "1"
		case "boolean":
			return "true"
		case "array":
			if schema.Items != nil && schema.Items.IsA() {
				itemExample := d.getExampleValue(schema.Items.A.Schema())
				return fmt.Sprintf("[%s]", itemExample)
			}
			return "[]"
		case "object":
			return "{}"
		}
	}

	return "null"
}
