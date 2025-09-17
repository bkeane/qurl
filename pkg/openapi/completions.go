package openapi

import (
	"context"
	"fmt"
)

// PathCompletions returns path completions for shell autocomplete
// Returns all paths matching the method filter - shell handles prefix filtering
func (v *Viewer) PathCompletions(ctx context.Context, pathPrefix, method string) ([]string, error) {
	if err := v.ensureSpecLoaded(ctx); err != nil {
		return nil, err
	}

	paths, err := v.parser.GetPaths("*", method)
	if err != nil {
		return nil, fmt.Errorf("getting paths: %w", err)
	}

	// Return all paths - let the shell do the filtering
	var completions []string
	for _, path := range paths {
		completions = append(completions, path.Path)
	}
	return completions, nil
}

// ParamCompletions returns parameter names for a specific path and method
func (v *Viewer) ParamCompletions(ctx context.Context, path, method string) ([]string, error) {
	if err := v.ensureSpecLoaded(ctx); err != nil {
		return nil, err
	}

	paths, err := v.parser.GetPaths(path, method)
	if err != nil {
		return nil, fmt.Errorf("getting paths: %w", err)
	}

	var paramNames []string
	paramSet := make(map[string]bool)

	for _, pathInfo := range paths {
		for _, param := range pathInfo.Parameters {
			// Only include query parameters for completion
			if param.In == "query" && !paramSet[param.Name] {
				paramNames = append(paramNames, param.Name)
				paramSet[param.Name] = true
			}
		}
	}

	return paramNames, nil
}

// MethodCompletions returns HTTP methods available for a specific path
func (v *Viewer) MethodCompletions(ctx context.Context, path string) ([]string, error) {
	if err := v.ensureSpecLoaded(ctx); err != nil {
		return nil, err
	}

	paths, err := v.parser.GetPaths(path, "*")
	if err != nil {
		return nil, fmt.Errorf("getting paths: %w", err)
	}

	var methods []string
	methodSet := make(map[string]bool)

	for _, pathInfo := range paths {
		// Only include methods for exact path matches
		if pathInfo.Path == path && !methodSet[pathInfo.Method] {
			methods = append(methods, pathInfo.Method)
			methodSet[pathInfo.Method] = true
		}
	}

	return methods, nil
}