package mcp

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestEstimateTokens(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "empty string",
			input:    "",
			expected: 0,
		},
		{
			name:     "simple text",
			input:    "hello world",
			expected: 2, // 11 chars / 4 = 2
		},
		{
			name:     "json object",
			input:    `{"key": "value"}`,
			expected: 4, // 16 chars / 4 = 4
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := estimateTokens(tt.input)
			if result != tt.expected {
				t.Errorf("estimateTokens() = %d, want %d", result, tt.expected)
			}
		})
	}
}

func TestFilterRegex(t *testing.T) {
	tests := []struct {
		name         string
		body         string
		pattern      string
		contextLines int
		wantMatches  int
		wantError    bool
		checkContent func(string) bool
	}{
		{
			name: "single match with context",
			body: `line 1
line 2
ERROR: something failed
line 4
line 5`,
			pattern:      "ERROR",
			contextLines: 1,
			wantMatches:  1,
			wantError:    false,
			checkContent: func(s string) bool {
				return strings.Contains(s, "Context Window") &&
					strings.Contains(s, "line 2") &&
					strings.Contains(s, "ERROR: something failed") &&
					strings.Contains(s, "line 4")
			},
		},
		{
			name: "multiple matches",
			body: `ERROR: first error
info line
ERROR: second error
more info`,
			pattern:      "ERROR",
			contextLines: 0,
			wantMatches:  2,
			wantError:    false,
			checkContent: func(s string) bool {
				return strings.Contains(s, "Context Window") &&
					strings.Contains(s, "ERROR: first error") &&
					strings.Contains(s, "ERROR: second error")
			},
		},
		{
			name: "no matches",
			body: `line 1
line 2
line 3`,
			pattern:      "ERROR",
			contextLines: 1,
			wantMatches:  0,
			wantError:    false,
			checkContent: func(s string) bool {
				return s == ""
			},
		},
		{
			name:         "invalid regex",
			body:         "some text",
			pattern:      "[invalid",
			contextLines: 1,
			wantMatches:  0,
			wantError:    true,
		},
		{
			name: "context at boundaries",
			body: `ERROR at start
line 2`,
			pattern:      "ERROR",
			contextLines: 5,
			wantMatches:  1,
			wantError:    false,
			checkContent: func(s string) bool {
				return strings.Contains(s, "Context Window") &&
					strings.Contains(s, "ERROR at start") &&
					strings.Contains(s, "line 2")
			},
		},
		{
			name: "regex with special characters",
			body: `status: failed
status: success
status: failed`,
			pattern:      `status:.*failed`,
			contextLines: 0,
			wantMatches:  2,
			wantError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := filterRegex(tt.body, tt.pattern, tt.contextLines)

			if tt.wantError {
				if err == nil {
					t.Errorf("filterRegex() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("filterRegex() unexpected error: %v", err)
				return
			}

			// Check metadata
			filterMeta := result.Meta["filter"].(map[string]interface{})
			if filterMeta["type"] != "regex" {
				t.Errorf("filter type = %v, want regex", filterMeta["type"])
			}
			if filterMeta["pattern"] != tt.pattern {
				t.Errorf("filter pattern = %v, want %v", filterMeta["pattern"], tt.pattern)
			}
			if filterMeta["total_matches"] != tt.wantMatches {
				t.Errorf("total_matches = %v, want %v", filterMeta["total_matches"], tt.wantMatches)
			}

			// Check tokens metadata
			tokens := result.Meta["tokens"].(map[string]interface{})
			if tokens["source"] != estimateTokens(tt.body) {
				t.Errorf("source tokens mismatch")
			}
			if tokens["returned"] != estimateTokens(result.Content) {
				t.Errorf("returned tokens mismatch")
			}

			// Check bytes metadata
			bytes := result.Meta["bytes"].(map[string]interface{})
			if bytes["source"] != len(tt.body) {
				t.Errorf("source bytes = %v, want %v", bytes["source"], len(tt.body))
			}
			if bytes["returned"] != len(result.Content) {
				t.Errorf("returned bytes = %v, want %v", bytes["returned"], len(result.Content))
			}

			// Check content
			if tt.checkContent != nil && !tt.checkContent(result.Content) {
				t.Errorf("content check failed. Content:\n%s", result.Content)
			}
		})
	}
}

func TestFilterJMESPath(t *testing.T) {
	tests := []struct {
		name         string
		body         string
		expression   string
		wantCount    int
		wantError    bool
		checkContent func(string) bool
	}{
		{
			name:       "simple array filter",
			body:       `{"items": [{"status": "failed"}, {"status": "success"}, {"status": "failed"}]}`,
			expression: "items[?status=='failed']",
			wantCount:  2,
			wantError:  false,
			checkContent: func(s string) bool {
				var result []map[string]interface{}
				json.Unmarshal([]byte(s), &result)
				return len(result) == 2 && result[0]["status"] == "failed"
			},
		},
		{
			name:       "projection",
			body:       `{"users": [{"id": 1, "name": "Alice"}, {"id": 2, "name": "Bob"}]}`,
			expression: "users[].name",
			wantCount:  2,
			wantError:  false,
			checkContent: func(s string) bool {
				var result []string
				json.Unmarshal([]byte(s), &result)
				return len(result) == 2 && result[0] == "Alice"
			},
		},
		{
			name:       "single object result",
			body:       `{"user": {"id": 1, "name": "Alice"}}`,
			expression: "user",
			wantCount:  1,
			wantError:  false,
			checkContent: func(s string) bool {
				var result map[string]interface{}
				json.Unmarshal([]byte(s), &result)
				return result["name"] == "Alice"
			},
		},
		{
			name:       "null result",
			body:       `{"items": []}`,
			expression: "items[?status=='missing']",
			wantCount:  0,
			wantError:  false,
			checkContent: func(s string) bool {
				return strings.TrimSpace(s) == "[]" || strings.TrimSpace(s) == "null"
			},
		},
		{
			name:       "invalid json",
			body:       `not json`,
			expression: "items",
			wantError:  true,
		},
		{
			name:       "invalid jmespath",
			body:       `{"items": []}`,
			expression: "items[invalid",
			wantError:  true,
		},
		{
			name:       "complex nested query",
			body:       `{"orders": [{"id": 1, "items": [{"price": 10}, {"price": 20}]}, {"id": 2, "items": [{"price": 30}]}]}`,
			expression: "orders[].items[].price",
			wantCount:  3,
			wantError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := filterJMESPath(tt.body, tt.expression)

			if tt.wantError {
				if err == nil {
					t.Errorf("filterJMESPath() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("filterJMESPath() unexpected error: %v", err)
				return
			}

			// Check metadata
			filterMeta := result.Meta["filter"].(map[string]interface{})
			if filterMeta["type"] != "jmespath" {
				t.Errorf("filter type = %v, want jmespath", filterMeta["type"])
			}
			if filterMeta["expression"] != tt.expression {
				t.Errorf("filter expression = %v, want %v", filterMeta["expression"], tt.expression)
			}
			if filterMeta["result_count"] != tt.wantCount {
				t.Errorf("result_count = %v, want %v", filterMeta["result_count"], tt.wantCount)
			}

			// Check tokens metadata
			tokens := result.Meta["tokens"].(map[string]interface{})
			if tokens["source"] != estimateTokens(tt.body) {
				t.Errorf("source tokens mismatch")
			}
			if tokens["returned"] != estimateTokens(result.Content) {
				t.Errorf("returned tokens mismatch")
			}

			// Check bytes metadata
			bytes := result.Meta["bytes"].(map[string]interface{})
			if bytes["source"] != len(tt.body) {
				t.Errorf("source bytes = %v, want %v", bytes["source"], len(tt.body))
			}

			// Check content
			if tt.checkContent != nil && !tt.checkContent(result.Content) {
				t.Errorf("content check failed. Content:\n%s", result.Content)
			}
		})
	}
}

func TestMaxMin(t *testing.T) {
	if max(5, 3) != 5 {
		t.Errorf("max(5, 3) = %d, want 5", max(5, 3))
	}
	if max(3, 5) != 5 {
		t.Errorf("max(3, 5) = %d, want 5", max(3, 5))
	}
	if min(5, 3) != 3 {
		t.Errorf("min(5, 3) = %d, want 3", min(5, 3))
	}
	if min(3, 5) != 3 {
		t.Errorf("min(3, 5) = %d, want 3", min(3, 5))
	}
}