package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/brendan.keane/qurl/internal/config"
	"github.com/brendan.keane/qurl/internal/errors"
	"github.com/brendan.keane/qurl/internal/http"
	"github.com/brendan.keane/qurl/pkg/openapi"
	"github.com/rs/zerolog"
)

const (
	discoverToolDescription = `Use this tool BEFORE calling execute to understand what endpoints are available, what parameters they accept, and what response structure they return.

**When to use:**
- Always call this FIRST when exploring a new API or endpoint
- Call with no parameters to browse all available endpoints
- Call with a specific path to see detailed documentation including:
  - Required and optional parameters
  - Request body schemas
  - Response body structure and JSON keys
  - This helps you craft targeted jmespath or regex filters to extract only the data you need

**Why this matters:**
Understanding the response structure lets you use filters effectively, saving tokens by requesting only relevant data instead of the full response.`

	executeToolDescription = `Make HTTP requests to API endpoints. IMPORTANT: Always use filters (jmespath or regex) when you only need specific data from the response.

**Best practices:**
1. Use 'discover' tool first to understand the response structure
2. Use 'jmespath' filter for JSON responses when you only need specific fields (saves tokens!)
3. Use 'regex' filter to search for specific patterns in any response type
4. Only request unfiltered responses when you need the complete data

**Filtering is strongly encouraged** - it reduces token usage and helps you extract exactly what you need. If the user asks for specific information, always try to filter the response rather than returning everything.`

	discoverPathParamDescription = `Path filter. Omit or use '*' to list all available endpoints.

Provide a specific path (e.g., /users/123) to see complete documentation including:
- Request parameters (query, path, header)
- Request body schemas with field descriptions
- Response body structure showing all available JSON keys
- This information is ESSENTIAL for crafting effective jmespath or regex filters`

	discoverMethodParamDescription = `HTTP method filter (GET, POST, PUT, DELETE, etc.).

Specify a method to see detailed schemas for that specific operation. Use 'ANY' or omit to see all available methods for the path.`

	executePathParamDescription      = `API endpoint path (required)`
	executeMethodParamDescription    = `HTTP method (GET, POST, PUT, DELETE, etc.)`
	executeHeadersParamDescription   = `HTTP headers as key-value pairs`
	executeQueryParamDescription     = `Query parameters as key-value pairs`
	executeBodyParamDescription      = `Request body data`
	executeRegexParamDescription     = `Regex pattern to search response text (returns matches with surrounding context).

Works with any text format including minified JSON. Use this when searching for specific patterns or terms. Cannot be used with jmespath.

**When to use:** Searching for specific strings, patterns, or when you don't know the exact JSON structure.`

	executeJMESPathParamDescription = `JMESPath expression to filter JSON response (https://jmespath.org).

**STRONGLY RECOMMENDED** when working with JSON responses and you only need specific fields. This dramatically reduces token usage.

Cannot be used with regex.

**Examples:**
- 'items[].name' - extract just the name field from an array
- 'data.{id: id, name: name}' - extract only id and name fields
- 'items[?price > 100]' - filter items by condition

**When to use:** Always prefer this for JSON responses when you need specific fields rather than the entire response.`

	executeContextLinesParamDescription = `Amount of context to show around regex matches. Multiplied by ~80 characters per 'line' (default: 5 = ~400 chars of context).

Only used with regex parameter. Increase for more context, decrease for more precise matches.`
)

// Server implements the MCP server protocol
type Server struct {
	logger     zerolog.Logger
	config     *config.Config
	executor   http.HTTPExecutor
	viewer     *openapi.Viewer
}

// MCPRequest represents an incoming MCP request
type MCPRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// MCPResponse represents an outgoing MCP response
type MCPResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *MCPError   `json:"error,omitempty"`
}

// MCPError represents an MCP error response
type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// NewServer creates a new MCP server
func NewServer(logger zerolog.Logger, cfg *config.Config) (*Server, error) {
	// Initialize HTTP client factory
	factory := http.NewClientFactory(logger)

	// Create HTTP executor for request processing
	executor, err := factory.CreateExecutor(cfg)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeNetwork, "failed to create HTTP executor for MCP server")
	}

	// Create OpenAPI viewer with authenticated HTTP client
	authClient := http.NewAuthenticatedHTTPClient(cfg, logger)
	viewer := openapi.NewViewer(authClient, cfg.OpenAPIURL)

	return &Server{
		logger:     logger.With().Str("component", "mcp_server").Logger(),
		config:     cfg,
		executor:   executor,
		viewer:     viewer,
	}, nil
}

// Start begins the MCP server message loop
func (s *Server) Start() error {
	s.logger.Debug().Msg("MCP server started, reading from stdin")

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		s.logger.Debug().Str("raw_message", line).Msg("received MCP message")

		if err := s.handleMessage(line); err != nil {
			s.logger.Error().Err(err).Msg("error handling MCP message")
			// Continue processing other messages
		}
	}

	if err := scanner.Err(); err != nil {
		s.logger.Error().Err(err).Msg("error reading from stdin")
		return errors.Wrap(err, errors.ErrorTypeInternal, "failed to read MCP messages from stdin")
	}

	s.logger.Debug().Msg("MCP server stopped")
	return nil
}

// handleMessage processes a single MCP message
func (s *Server) handleMessage(line string) error {
	var req MCPRequest
	if err := json.Unmarshal([]byte(line), &req); err != nil {
		s.logger.Warn().Err(err).Str("line", line).Msg("failed to parse MCP message")
		return s.sendError(nil, -32700, "Parse error")
	}

	s.logger.Debug().
		Interface("id", req.ID).
		Str("method", req.Method).
		Msg("processing MCP request")

	switch req.Method {
	case "initialize":
		return s.handleInitialize(req.ID, req.Params)
	case "tools/list":
		return s.handleToolsList(req.ID)
	case "tools/call":
		return s.handleToolsCall(req.ID, req.Params)
	case "notifications/cancelled":
		// This is a notification, no response needed
		s.logger.Debug().Msg("received cancellation notification")
		return nil
	default:
		s.logger.Warn().Str("method", req.Method).Msg("unknown MCP method")
		return s.sendError(req.ID, -32601, fmt.Sprintf("Method not found: %s", req.Method))
	}
}

// handleInitialize handles the MCP initialize request
func (s *Server) handleInitialize(id interface{}, params interface{}) error {
	// Build server info with optional description
	serverInfo := map[string]interface{}{
		"name":    "qurl",
		"version": "1.0.0",
	}

	// Add description if configured
	if s.config.MCP.Description != "" {
		serverInfo["description"] = s.config.MCP.Description
	}

	// MCP initialize response with server capabilities
	response := MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"serverInfo":      serverInfo,
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{},
			},
		},
	}

	s.logger.Debug().
		Str("description", s.config.MCP.Description).
		Msg("sent MCP initialize response")
	return s.sendResponse(response)
}

// handleToolsList returns the list of available tools
func (s *Server) handleToolsList(id interface{}) error {
	tools := []map[string]interface{}{
		{
			"name":        "discover",
			"description": discoverToolDescription,
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": discoverPathParamDescription,
					},
					"method": map[string]interface{}{
						"type":        "string",
						"description": discoverMethodParamDescription,
					},
				},
			},
		},
		{
			"name":        "execute",
			"description": executeToolDescription,
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": executePathParamDescription,
					},
					"method": map[string]interface{}{
						"type":        "string",
						"description": executeMethodParamDescription,
						"default":     "GET",
					},
					"headers": map[string]interface{}{
						"type":        "object",
						"description": executeHeadersParamDescription,
					},
					"query": map[string]interface{}{
						"type":        "object",
						"description": executeQueryParamDescription,
					},
					"body": map[string]interface{}{
						"type":        "string",
						"description": executeBodyParamDescription,
					},
					"regex": map[string]interface{}{
						"type":        "string",
						"description": executeRegexParamDescription,
					},
					"jmespath": map[string]interface{}{
						"type":        "string",
						"description": executeJMESPathParamDescription,
					},
					"context_lines": map[string]interface{}{
						"type":        "integer",
						"description": executeContextLinesParamDescription,
						"default":     5,
					},
				},
				"required": []string{"path"},
			},
		},
	}

	response := MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result: map[string]interface{}{
			"tools": tools,
		},
	}

	return s.sendResponse(response)
}

// handleToolsCall executes a tool call
func (s *Server) handleToolsCall(id interface{}, params interface{}) error {
	paramsMap, ok := params.(map[string]interface{})
	if !ok {
		return s.sendError(id, -32602, "Invalid params")
	}

	name, ok := paramsMap["name"].(string)
	if !ok {
		return s.sendError(id, -32602, "Missing tool name")
	}

	arguments, ok := paramsMap["arguments"].(map[string]interface{})
	if !ok {
		arguments = make(map[string]interface{})
	}

	s.logger.Debug().
		Str("tool_name", name).
		Interface("arguments", arguments).
		Msg("executing tool call")

	switch name {
	case "discover":
		return s.executeDiscover(id, arguments)
	case "execute":
		return s.executeHTTPRequest(id, arguments)
	default:
		return s.sendError(id, -32601, fmt.Sprintf("Unknown tool: %s", name))
	}
}

// executeDiscover handles the discover tool
func (s *Server) executeDiscover(id interface{}, args map[string]interface{}) error {
	path := ""
	if p, ok := args["path"].(string); ok {
		path = p
	}

	method := ""
	if m, ok := args["method"].(string); ok {
		method = m
	}

	// Apply path prefix constraint
	if s.config.MCP.PathPrefix != "" {
		if path == "" || path == "*" {
			// If no path specified, use the prefix
			path = s.config.MCP.PathPrefix + "*"
		} else if !strings.HasPrefix(path, s.config.MCP.PathPrefix) {
			// If path doesn't start with prefix, prepend it
			path = s.config.MCP.PathPrefix + strings.TrimPrefix(path, "/")
		}
	}

	// Apply method constraints
	if len(s.config.MCP.AllowedMethods) > 0 {
		if method != "" && method != "ANY" {
			// Check if requested method is allowed
			methodAllowed := false
			for _, allowed := range s.config.MCP.AllowedMethods {
				if strings.EqualFold(allowed, method) {
					methodAllowed = true
					break
				}
			}
			if !methodAllowed {
				s.logger.Warn().
					Str("method", method).
					Strs("allowed", s.config.MCP.AllowedMethods).
					Msg("requested method not in allowed list")
				return s.sendError(id, -32602, fmt.Sprintf("Method %s not allowed. Allowed methods: %v",
					method, s.config.MCP.AllowedMethods))
			}
		} else {
			// If no method specified, show only allowed methods
			if len(s.config.MCP.AllowedMethods) == 1 {
				method = s.config.MCP.AllowedMethods[0]
			} else {
				// For multiple allowed methods, we'll filter the output
				method = "ANY"
			}
		}
	}

	s.logger.Debug().
		Str("discover_path", path).
		Str("discover_method", method).
		Strs("allowed_methods", s.config.MCP.AllowedMethods).
		Str("path_prefix", s.config.MCP.PathPrefix).
		Msg("discovering endpoints with constraints")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Use the OpenAPI viewer to get documentation
	if path == "" {
		path = "*"
	}
	if method == "" {
		method = "ANY"
	}

	output, err := s.viewer.View(ctx, path, method)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to discover endpoints")
		return s.sendError(id, -32603, fmt.Sprintf("Failed to discover endpoints: %v", err))
	}

	// TODO: Filter output to only show allowed methods if multiple are configured
	// This would require parsing the output and filtering, which is more complex

	response := MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result: map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": output,
				},
			},
		},
	}

	return s.sendResponse(response)
}

// executeHTTPRequest handles the execute tool
func (s *Server) executeHTTPRequest(id interface{}, args map[string]interface{}) error {
	path, ok := args["path"].(string)
	if !ok || path == "" {
		return s.sendError(id, -32602, "Missing required parameter: path")
	}

	method := "GET"
	if m, ok := args["method"].(string); ok && m != "" {
		method = strings.ToUpper(m)
	}

	// Validate path constraint
	if s.config.MCP.PathPrefix != "" {
		if !strings.HasPrefix(path, s.config.MCP.PathPrefix) {
			s.logger.Warn().
				Str("path", path).
				Str("prefix", s.config.MCP.PathPrefix).
				Msg("path outside allowed prefix")
			return s.sendError(id, -32602, fmt.Sprintf("Path %s not allowed. Must be under %s",
				path, s.config.MCP.PathPrefix))
		}
	}

	// Validate method constraint
	if len(s.config.MCP.AllowedMethods) > 0 {
		methodAllowed := false
		for _, allowed := range s.config.MCP.AllowedMethods {
			if strings.EqualFold(allowed, method) {
				methodAllowed = true
				break
			}
		}
		if !methodAllowed {
			s.logger.Warn().
				Str("method", method).
				Strs("allowed", s.config.MCP.AllowedMethods).
				Msg("method not in allowed list")
			return s.sendError(id, -32602, fmt.Sprintf("Method %s not allowed. Allowed methods: %v",
				method, s.config.MCP.AllowedMethods))
		}
	}

	// Create a temporary config for this request
	requestConfig := *s.config
	requestConfig.Methods = []string{method}

	// Start with inherited headers from MCP config
	requestConfig.Headers = append([]string{}, s.config.MCP.Headers...)

	// Add request-specific headers (these can override inherited ones)
	if headers, ok := args["headers"].(map[string]interface{}); ok {
		for key, value := range headers {
			if strValue, ok := value.(string); ok {
				requestConfig.Headers = append(requestConfig.Headers, fmt.Sprintf("%s: %s", key, strValue))
			}
		}
	}

	// Handle query parameters
	if query, ok := args["query"].(map[string]interface{}); ok {
		requestConfig.QueryParams = []string{}
		for key, value := range query {
			if strValue, ok := value.(string); ok {
				requestConfig.QueryParams = append(requestConfig.QueryParams, fmt.Sprintf("%s=%s", key, strValue))
			}
		}
	}

	// Handle body
	if body, ok := args["body"].(string); ok {
		requestConfig.Data = body
	}

	s.logger.Debug().
		Str("method", method).
		Str("path", path).
		Int("headers", len(requestConfig.Headers)).
		Int("query_params", len(requestConfig.QueryParams)).
		Bool("has_body", requestConfig.Data != "").
		Msg("executing HTTP request via MCP")

	// Create a new HTTP client with the request-specific config
	factory := http.NewClientFactory(s.logger)
	executor, err := factory.CreateExecutor(&requestConfig)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to create HTTP executor for MCP request")
		return s.sendError(id, -32603, fmt.Sprintf("Failed to create HTTP executor: %v", err))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Execute the request and capture response
	body, headers, statusCode, err := executor.ExecuteForMCP(ctx, path)
	if err != nil {
		s.logger.Error().Err(err).Msg("HTTP request failed via MCP")
		return s.sendError(id, -32603, fmt.Sprintf("HTTP request failed: %v", err))
	}

	// Check for filter parameters
	regexPattern, hasRegex := args["regex"].(string)
	jmespathExpr, hasJMESPath := args["jmespath"].(string)

	// Empty strings should not trigger filtering
	if hasRegex && strings.TrimSpace(regexPattern) == "" {
		hasRegex = false
	}
	if hasJMESPath && strings.TrimSpace(jmespathExpr) == "" {
		hasJMESPath = false
	}

	// Validate mutual exclusivity
	if hasRegex && hasJMESPath {
		return s.sendError(id, -32602, "Cannot use both regex and jmespath filters simultaneously")
	}

	// Apply regex filter if requested
	if hasRegex {
		contextLines := 5 // default
		if cl, ok := args["context_lines"].(float64); ok {
			contextLines = int(cl)
		}

		s.logger.Debug().
			Str("pattern", regexPattern).
			Int("context_lines", contextLines).
			Msg("applying regex filter")

		filterResult, err := filterRegex(body, regexPattern, contextLines)
		if err != nil {
			s.logger.Error().Err(err).Msg("regex filter failed")
			return s.sendError(id, -32603, fmt.Sprintf("Regex filter failed: %v", err))
		}

		return s.sendFilteredResponse(id, filterResult, statusCode, headers, &requestConfig)
	}

	// Apply jmespath filter if requested
	if hasJMESPath {
		s.logger.Debug().
			Str("expression", jmespathExpr).
			Msg("applying jmespath filter")

		filterResult, err := filterJMESPath(body, jmespathExpr)
		if err != nil {
			s.logger.Error().Err(err).Msg("jmespath filter failed")
			return s.sendError(id, -32603, fmt.Sprintf("JMESPath filter failed: %v", err))
		}

		return s.sendFilteredResponse(id, filterResult, statusCode, headers, &requestConfig)
	}

	// No filtering - return raw response
	// Format response with status code and headers if verbose
	var responseText string
	if requestConfig.Verbose || requestConfig.IncludeHeaders {
		// Build a formatted response similar to CLI output
		responseText = fmt.Sprintf("HTTP Status: %d\n", statusCode)
		if requestConfig.IncludeHeaders {
			responseText += "\nHeaders:\n"
			for key, values := range headers {
				for _, value := range values {
					responseText += fmt.Sprintf("%s: %s\n", key, value)
				}
			}
			responseText += "\n"
		}
		responseText += body
	} else {
		responseText = body
	}

	response := MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result: map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": responseText,
				},
			},
		},
	}

	return s.sendResponse(response)
}

// sendFilteredResponse sends an MCP response with filtered content and metadata
func (s *Server) sendFilteredResponse(id interface{}, filterResult *FilterResult, statusCode int, headers map[string][]string, cfg *config.Config) error {
	contentText := filterResult.Content

	// Prepend status/headers if verbose mode is enabled
	if cfg.Verbose || cfg.IncludeHeaders {
		prefix := fmt.Sprintf("HTTP Status: %d\n", statusCode)
		if cfg.IncludeHeaders {
			prefix += "\nHeaders:\n"
			for key, values := range headers {
				for _, value := range values {
					prefix += fmt.Sprintf("%s: %s\n", key, value)
				}
			}
			prefix += "\n"
		}
		contentText = prefix + contentText
	}

	response := MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result: map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": contentText,
				},
			},
			"_meta": filterResult.Meta,
		},
	}

	return s.sendResponse(response)
}

// sendResponse sends an MCP response
func (s *Server) sendResponse(response MCPResponse) error {
	data, err := json.Marshal(response)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to marshal MCP response")
		return err
	}

	fmt.Println(string(data))
	s.logger.Debug().
		Interface("response_id", response.ID).
		Msg("sent MCP response")

	return nil
}

// sendError sends an MCP error response
func (s *Server) sendError(id interface{}, code int, message string) error {
	response := MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &MCPError{
			Code:    code,
			Message: message,
		},
	}

	return s.sendResponse(response)
}