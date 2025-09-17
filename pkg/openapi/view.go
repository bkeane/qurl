package openapi

import (
	"context"
	"fmt"
	"strings"
)

type Viewer struct {
	parser    *Parser
	displayer *Displayer
	specURL   string
}

func NewViewer(client HTTPClient, specURL string) *Viewer {
	parser := NewParserWithClient(client)
	return &Viewer{
		parser:    parser,
		displayer: NewDisplayer(parser),
		specURL:   specURL,
	}
}

// ensureSpecLoaded loads the OpenAPI spec if it hasn't been loaded yet
func (v *Viewer) ensureSpecLoaded(ctx context.Context) error {
	if v.specURL != "" && v.parser.model == nil {
		if err := v.parser.LoadFromURL(ctx, v.specURL); err != nil {
			return fmt.Errorf("loading OpenAPI spec: %w", err)
		}
	}
	return nil
}

func (v *Viewer) View(ctx context.Context, path, method string) (string, error) {
	// Check if user wants index view (trailing slash)
	showIndex := strings.HasSuffix(path, "/")
	if showIndex {
		// Convert trailing slash to wildcard pattern to get all sub-paths
		path = strings.TrimSuffix(path, "/") + "*"
	}

	if err := v.ensureSpecLoaded(ctx); err != nil {
		return "", err
	}

	paths, err := v.parser.GetPaths(path, method)
	if err != nil {
		return "", fmt.Errorf("getting paths: %w", err)
	}

	if len(paths) == 0 {
		return "No endpoints found matching the specified path and method", nil
	}

	// If trailing slash or wildcard, show index
	if showIndex || path == "" || path == "*" {
		return v.displayer.RenderIndex(paths), nil
	}

	// Find exact matches for the path
	var exactMatches []PathInfo
	for _, p := range paths {
		if p.Path == path {
			exactMatches = append(exactMatches, p)
		}
	}

	// If we have exact matches, show them
	if len(exactMatches) == 1 {
		return v.displayer.RenderOperation(exactMatches[0]), nil
	} else if len(exactMatches) > 0 {
		// Multiple methods for the same exact path
		return v.displayer.RenderIndex(exactMatches), nil
	}

	// No exact match, show all matching paths as index
	return v.displayer.RenderIndex(paths), nil
}

func (v *Viewer) ViewFromBytes(data []byte, path, method string) (string, error) {
	// Check if user wants index view (trailing slash)
	showIndex := strings.HasSuffix(path, "/")
	if showIndex {
		path = strings.TrimSuffix(path, "/")
	}

	if err := v.parser.LoadFromBytes(data); err != nil {
		return "", fmt.Errorf("loading OpenAPI spec: %w", err)
	}

	paths, err := v.parser.GetPaths(path, method)
	if err != nil {
		return "", fmt.Errorf("getting paths: %w", err)
	}

	if len(paths) == 0 {
		return "No endpoints found matching the specified path and method", nil
	}

	// If trailing slash or wildcard, show index
	if showIndex || path == "" || path == "*" {
		return v.displayer.RenderIndex(paths), nil
	}

	// Find exact matches for the path
	var exactMatches []PathInfo
	for _, p := range paths {
		if p.Path == path {
			exactMatches = append(exactMatches, p)
		}
	}

	// If we have exact matches, show them
	if len(exactMatches) == 1 {
		return v.displayer.RenderOperation(exactMatches[0]), nil
	} else if len(exactMatches) > 0 {
		// Multiple methods for the same exact path
		return v.displayer.RenderIndex(exactMatches), nil
	}

	// No exact match, show all matching paths as index
	return v.displayer.RenderIndex(paths), nil
}