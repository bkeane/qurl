package openapi

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
)

// HTTPClient interface for making HTTP requests
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Parser struct {
	document   libopenapi.Document
	model      *libopenapi.DocumentModel[v3.Document]
	httpClient HTTPClient
}

func NewParser() *Parser {
	return &Parser{
		httpClient: http.DefaultClient,
	}
}

func NewParserWithClient(client HTTPClient) *Parser {
	return &Parser{
		httpClient: client,
	}
}

func (p *Parser) LoadFromURL(ctx context.Context, urlStr string) error {
	// Parse URL to check scheme
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("parsing URL: %w", err)
	}

	// Handle file:// URIs by reading from local filesystem
	if parsedURL.Scheme == "file" {
		filePath := parsedURL.Path

		// Special case: file://host/path URLs where host is not empty
		// This means the path is actually host + path, treating it as relative
		if parsedURL.Host != "" {
			filePath = parsedURL.Host + parsedURL.Path
		}

		// Handle relative paths - if path doesn't start with /, treat as relative to current directory
		if !filepath.IsAbs(filePath) {
			var err error
			filePath, err = filepath.Abs(filePath)
			if err != nil {
				return fmt.Errorf("resolving absolute path: %w", err)
			}
		}
		return p.loadFromFile(filePath)
	}

	// Handle HTTP/HTTPS/Lambda URIs
	req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("fetching OpenAPI spec: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %w", err)
	}

	return p.LoadFromBytes(body)
}

// loadFromFile loads an OpenAPI specification from a local file
func (p *Parser) loadFromFile(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("reading file %s: %w", filePath, err)
	}

	return p.LoadFromBytes(data)
}

func (p *Parser) LoadFromBytes(data []byte) error {
	document, err := libopenapi.NewDocument(data)
	if err != nil {
		return fmt.Errorf("parsing OpenAPI document: %w", err)
	}

	model, errs := document.BuildV3Model()
	if len(errs) > 0 {
		return fmt.Errorf("building v3 model: %v", errs)
	}

	p.document = document
	p.model = model

	return nil
}

type PathInfo struct {
	Path        string
	Method      string
	Summary     string
	Description string
	Operation   *v3.Operation
	Parameters  []*v3.Parameter
	RequestBody *v3.RequestBody
	Responses   *v3.Responses
}

func (p *Parser) GetPaths(pathFilter, methodFilter string) ([]PathInfo, error) {
	if p.model == nil {
		return nil, fmt.Errorf("no OpenAPI document loaded")
	}

	var paths []PathInfo

	pathItems := p.model.Model.Paths.PathItems
	if pathItems == nil {
		return paths, nil
	}

	for pathPattern, pathItem := range pathItems.FromOldest() {
		if !matchesPathFilter(pathPattern, pathFilter) {
			continue
		}

		operations := getOperations(pathItem)
		for method, op := range operations {
			if !matchesMethodFilter(method, methodFilter) {
				continue
			}

			info := PathInfo{
				Path:        pathPattern,
				Method:      strings.ToUpper(method),
				Operation:   op,
				Parameters:  mergeParameters(pathItem.Parameters, op.Parameters),
				RequestBody: op.RequestBody,
				Responses:   op.Responses,
			}

			if op.Summary != "" {
				info.Summary = op.Summary
			}
			if op.Description != "" {
				info.Description = op.Description
			}

			paths = append(paths, info)
		}
	}

	sort.Slice(paths, func(i, j int) bool {
		if paths[i].Path != paths[j].Path {
			return paths[i].Path < paths[j].Path
		}
		return methodOrder(paths[i].Method) < methodOrder(paths[j].Method)
	})

	return paths, nil
}

func (p *Parser) GetInfo() (*base.Info, error) {
	if p.model == nil {
		return nil, fmt.Errorf("no OpenAPI document loaded")
	}
	return p.model.Model.Info, nil
}

func (p *Parser) GetServers() ([]*v3.Server, error) {
	if p.model == nil {
		return nil, fmt.Errorf("no OpenAPI document loaded")
	}
	return p.model.Model.Servers, nil
}

func (p *Parser) GetSecuritySchemes() (*orderedmap.Map[string, *v3.SecurityScheme], error) {
	if p.model == nil {
		return nil, fmt.Errorf("no OpenAPI document loaded")
	}
	if p.model.Model.Components == nil {
		return nil, nil
	}
	return p.model.Model.Components.SecuritySchemes, nil
}

func (p *Parser) GetSecurity() []*base.SecurityRequirement {
	if p.model == nil {
		return nil
	}
	return p.model.Model.Security
}

func (p *Parser) GetTags() ([]*base.Tag, error) {
	if p.model == nil {
		return nil, fmt.Errorf("no OpenAPI document loaded")
	}
	return p.model.Model.Tags, nil
}

func matchesPathFilter(path, filter string) bool {
	if filter == "" || filter == "*" {
		return true
	}

	if strings.HasSuffix(filter, "*") {
		prefix := strings.TrimSuffix(filter, "*")
		return strings.HasPrefix(path, prefix)
	}

	// Check for exact match only
	return path == filter
}

func matchesMethodFilter(method, filter string) bool {
	if filter == "" || strings.EqualFold(filter, "ANY") || filter == "*" {
		return true
	}

	// Handle multiple methods separated by commas
	if strings.Contains(filter, ",") {
		methods := strings.Split(filter, ",")
		for _, m := range methods {
			if strings.EqualFold(method, strings.TrimSpace(m)) {
				return true
			}
		}
		return false
	}

	return strings.EqualFold(method, filter)
}

func getOperations(pathItem *v3.PathItem) map[string]*v3.Operation {
	ops := make(map[string]*v3.Operation)

	if pathItem.Get != nil {
		ops["get"] = pathItem.Get
	}
	if pathItem.Post != nil {
		ops["post"] = pathItem.Post
	}
	if pathItem.Put != nil {
		ops["put"] = pathItem.Put
	}
	if pathItem.Delete != nil {
		ops["delete"] = pathItem.Delete
	}
	if pathItem.Patch != nil {
		ops["patch"] = pathItem.Patch
	}
	if pathItem.Head != nil {
		ops["head"] = pathItem.Head
	}
	if pathItem.Options != nil {
		ops["options"] = pathItem.Options
	}

	return ops
}

func mergeParameters(pathParams, opParams []*v3.Parameter) []*v3.Parameter {
	paramMap := make(map[string]*v3.Parameter)

	for _, p := range pathParams {
		if p.Name != "" && p.In != "" {
			key := fmt.Sprintf("%s:%s", p.In, p.Name)
			paramMap[key] = p
		}
	}

	for _, p := range opParams {
		if p.Name != "" && p.In != "" {
			key := fmt.Sprintf("%s:%s", p.In, p.Name)
			paramMap[key] = p
		}
	}

	var result []*v3.Parameter
	for _, p := range paramMap {
		result = append(result, p)
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].In != result[j].In {
			return parameterInOrder(result[i].In) < parameterInOrder(result[j].In)
		}
		return result[i].Name < result[j].Name
	})

	return result
}

func methodOrder(method string) int {
	order := map[string]int{
		"GET":     0,
		"POST":    1,
		"PUT":     2,
		"PATCH":   3,
		"DELETE":  4,
		"HEAD":    5,
		"OPTIONS": 6,
	}
	if v, ok := order[method]; ok {
		return v
	}
	return 999
}

func parameterInOrder(in string) int {
	order := map[string]int{
		"path":   0,
		"query":  1,
		"header": 2,
		"cookie": 3,
	}
	if v, ok := order[in]; ok {
		return v
	}
	return 999
}
