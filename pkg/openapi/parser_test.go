package openapi

import (
	"context"
	"testing"
	"time"
)

const petstoreURL = "https://petstore3.swagger.io/api/v3/openapi.json"

func TestParserLoadFromURL(t *testing.T) {
	parser := NewParser()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := parser.LoadFromURL(ctx, petstoreURL)
	if err != nil {
		t.Fatalf("Failed to load OpenAPI spec from URL: %v", err)
	}

	info, err := parser.GetInfo()
	if err != nil {
		t.Fatalf("Failed to get info: %v", err)
	}

	if info.Title != "Swagger Petstore - OpenAPI 3.0" {
		t.Errorf("Expected title 'Swagger Petstore - OpenAPI 3.0', got '%s'", info.Title)
	}
}

func TestParserGetPaths(t *testing.T) {
	parser := NewParser()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := parser.LoadFromURL(ctx, petstoreURL)
	if err != nil {
		t.Fatalf("Failed to load OpenAPI spec: %v", err)
	}

	tests := []struct {
		name         string
		pathFilter   string
		methodFilter string
		wantMinPaths int
		checkPath    string
		checkMethod  string
	}{
		{
			name:         "Get all paths",
			pathFilter:   "*",
			methodFilter: "*",
			wantMinPaths: 10,
		},
		{
			name:         "Get specific path",
			pathFilter:   "/pet/{petId}",
			methodFilter: "*",
			wantMinPaths: 3,
		},
		{
			name:         "Get specific method",
			pathFilter:   "/pet/{petId}",
			methodFilter: "GET",
			wantMinPaths: 1,
			checkPath:    "/pet/{petId}",
			checkMethod:  "GET",
		},
		{
			name:         "Filter by path prefix",
			pathFilter:   "/pet*",
			methodFilter: "*",
			wantMinPaths: 5,
		},
		{
			name:         "Filter POST methods only",
			pathFilter:   "*",
			methodFilter: "POST",
			wantMinPaths: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paths, err := parser.GetPaths(tt.pathFilter, tt.methodFilter)
			if err != nil {
				t.Fatalf("Failed to get paths: %v", err)
			}

			if len(paths) < tt.wantMinPaths {
				t.Errorf("Expected at least %d paths, got %d", tt.wantMinPaths, len(paths))
			}

			if tt.checkPath != "" && tt.checkMethod != "" {
				found := false
				for _, path := range paths {
					if path.Path == tt.checkPath && path.Method == tt.checkMethod {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected to find path %s with method %s", tt.checkPath, tt.checkMethod)
				}
			}
		})
	}
}

func TestParserGetServers(t *testing.T) {
	parser := NewParser()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := parser.LoadFromURL(ctx, petstoreURL)
	if err != nil {
		t.Fatalf("Failed to load OpenAPI spec: %v", err)
	}

	servers, err := parser.GetServers()
	if err != nil {
		t.Fatalf("Failed to get servers: %v", err)
	}

	if len(servers) == 0 {
		t.Error("Expected at least one server")
	}
}

func TestMatchesPathFilter(t *testing.T) {
	tests := []struct {
		path   string
		filter string
		want   bool
	}{
		{"/pet/{petId}", "*", true},
		{"/pet/{petId}", "", true},
		{"/pet/{petId}", "/pet/{petId}", true},
		{"/pet/{petId}", "/pet*", true},
		{"/pet/{petId}", "/user*", false},
		{"/store/order", "/store*", true},
		{"/store/order", "/store/order", true},
		{"/store/order", "/store", false},
	}

	for _, tt := range tests {
		t.Run(tt.path+"_"+tt.filter, func(t *testing.T) {
			got := matchesPathFilter(tt.path, tt.filter)
			if got != tt.want {
				t.Errorf("matchesPathFilter(%q, %q) = %v, want %v", tt.path, tt.filter, got, tt.want)
			}
		})
	}
}

func TestMatchesMethodFilter(t *testing.T) {
	tests := []struct {
		method string
		filter string
		want   bool
	}{
		{"get", "*", true},
		{"GET", "*", true},
		{"get", "", true},
		{"get", "GET", true},
		{"GET", "get", true},
		{"post", "GET", false},
		{"POST", "get", false},
	}

	for _, tt := range tests {
		t.Run(tt.method+"_"+tt.filter, func(t *testing.T) {
			got := matchesMethodFilter(tt.method, tt.filter)
			if got != tt.want {
				t.Errorf("matchesMethodFilter(%q, %q) = %v, want %v", tt.method, tt.filter, got, tt.want)
			}
		})
	}
}

func TestMethodOrder(t *testing.T) {
	methods := []string{"DELETE", "POST", "GET", "PUT", "PATCH", "OPTIONS", "HEAD"}
	expectedOrder := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}

	for i := 0; i < len(methods)-1; i++ {
		for j := i + 1; j < len(methods); j++ {
			order1 := methodOrder(methods[i])
			order2 := methodOrder(methods[j])

			expectedIdx1 := indexOf(expectedOrder, methods[i])
			expectedIdx2 := indexOf(expectedOrder, methods[j])

			if expectedIdx1 < expectedIdx2 && order1 >= order2 {
				t.Errorf("Expected %s to come before %s", methods[i], methods[j])
			}
			if expectedIdx1 > expectedIdx2 && order1 <= order2 {
				t.Errorf("Expected %s to come after %s", methods[i], methods[j])
			}
		}
	}
}

func indexOf(slice []string, item string) int {
	for i, v := range slice {
		if v == item {
			return i
		}
	}
	return -1
}
