package mcp

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/jmespath/go-jmespath"
	"github.com/rs/zerolog/log"
)

// FilterResult represents the result of a filtering operation
type FilterResult struct {
	Content string                 `json:"content"`
	Meta    map[string]interface{} `json:"_meta"`
}

// estimateTokens approximates token count using chars/4 heuristic
func estimateTokens(data string) int {
	return len(data) / 4
}

// filterRegex searches text using regex and returns matches with context characters
func filterRegex(body string, pattern string, contextLines int) (*FilterResult, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regex pattern: %w", err)
	}

	// Convert context_lines to context characters (approximate: 80 chars per line)
	contextChars := contextLines * 80
	if contextChars < 100 {
		contextChars = 100 // minimum context
	}

	log.Debug().
		Int("input_bytes", len(body)).
		Str("pattern", pattern).
		Int("context_lines", contextLines).
		Int("context_chars", contextChars).
		Msg("filterRegex: starting")

	// Find all matches
	matches := re.FindAllStringIndex(body, -1)
	matchCount := len(matches)

	if matchCount == 0 {
		log.Debug().Msg("filterRegex: no matches found")
		return &FilterResult{
			Content: "",
			Meta: map[string]interface{}{
				"filter": map[string]interface{}{
					"type":          "regex",
					"pattern":       pattern,
					"total_matches": 0,
				},
				"tokens": map[string]interface{}{
					"returned": 0,
					"source":   estimateTokens(body),
				},
				"bytes": map[string]interface{}{
					"returned": 0,
					"source":   len(body),
				},
			},
		}, nil
	}

	log.Debug().
		Int("total_matches", matchCount).
		Msg("filterRegex: found matches")

	// Build context windows for each match
	type contextWindow struct {
		start int
		end   int
	}
	var windows []contextWindow

	for i, match := range matches {
		matchStart := match[0]
		matchEnd := match[1]

		// Calculate context window
		windowStart := max(0, matchStart-contextChars)
		windowEnd := min(len(body), matchEnd+contextChars)

		log.Debug().
			Int("match_num", i+1).
			Int("match_start", matchStart).
			Int("match_end", matchEnd).
			Int("window_start", windowStart).
			Int("window_end", windowEnd).
			Str("matched_text", body[matchStart:matchEnd]).
			Msg("filterRegex: processing match")

		windows = append(windows, contextWindow{start: windowStart, end: windowEnd})
	}

	// Merge overlapping windows
	merged := []contextWindow{windows[0]}
	for i := 1; i < len(windows); i++ {
		curr := windows[i]
		last := &merged[len(merged)-1]

		if curr.start <= last.end {
			// Windows overlap - merge them
			last.end = max(last.end, curr.end)
			log.Debug().
				Int("merged_start", last.start).
				Int("merged_end", last.end).
				Msg("filterRegex: merged overlapping windows")
		} else {
			merged = append(merged, curr)
		}
	}

	log.Debug().
		Int("original_windows", len(windows)).
		Int("merged_windows", len(merged)).
		Msg("filterRegex: finished merging")

	// Build output from merged windows
	var matchBlocks []string
	for i, window := range merged {
		excerpt := body[window.start:window.end]

		// Add ellipsis if we're not at the boundaries
		if window.start > 0 {
			excerpt = "..." + excerpt
		}
		if window.end < len(body) {
			excerpt = excerpt + "..."
		}

		header := fmt.Sprintf("=== Context Window %d (bytes %d-%d) ===", i+1, window.start, window.end)
		block := header + "\n" + excerpt

		matchBlocks = append(matchBlocks, block)
	}

	resultContent := strings.Join(matchBlocks, "\n\n")

	log.Debug().
		Int("result_bytes", len(resultContent)).
		Int("input_bytes", len(body)).
		Float64("reduction_pct", 100.0*float64(len(body)-len(resultContent))/float64(len(body))).
		Msg("filterRegex: result created")

	return &FilterResult{
		Content: resultContent,
		Meta: map[string]interface{}{
			"filter": map[string]interface{}{
				"type":          "regex",
				"pattern":       pattern,
				"total_matches": matchCount,
				"merged_windows": len(merged),
			},
			"tokens": map[string]interface{}{
				"returned": estimateTokens(resultContent),
				"source":   estimateTokens(body),
			},
			"bytes": map[string]interface{}{
				"returned": len(resultContent),
				"source":   len(body),
			},
		},
	}, nil
}

// filterJMESPath filters JSON using a JMESPath expression
func filterJMESPath(body string, expression string) (*FilterResult, error) {
	var data interface{}
	if err := json.Unmarshal([]byte(body), &data); err != nil {
		return nil, fmt.Errorf("invalid JSON response: %w", err)
	}

	result, err := jmespath.Search(expression, data)
	if err != nil {
		return nil, fmt.Errorf("invalid jmespath expression: %w", err)
	}

	filteredJSON, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal filtered result: %w", err)
	}

	// Count results
	resultCount := 0
	if arr, ok := result.([]interface{}); ok {
		resultCount = len(arr)
	} else if result != nil {
		resultCount = 1
	}

	resultContent := string(filteredJSON)

	return &FilterResult{
		Content: resultContent,
		Meta: map[string]interface{}{
			"filter": map[string]interface{}{
				"type":         "jmespath",
				"expression":   expression,
				"result_count": resultCount,
			},
			"tokens": map[string]interface{}{
				"returned": estimateTokens(resultContent),
				"source":   estimateTokens(body),
			},
			"bytes": map[string]interface{}{
				"returned": len(filteredJSON),
				"source":   len(body),
			},
		},
	}, nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}