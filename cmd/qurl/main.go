package main

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/config"
	qurlhttp "github.com/brendan.keane/qurl/pkg/http"
	"github.com/brendan.keane/qurl/pkg/openapi"
	"github.com/spf13/cobra"
)

var (
	openAPIURL     string
	httpMethod     string
	showDocs       bool
	queryParams    []string
	headers        []string
	verbose        bool
	includeHeaders bool
	data           string
	server         string
	sigV4Service   string
	sigV4Enabled   bool
	bearerToken    string
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

var rootCmd *cobra.Command

// getOpenAPIURL retrieves the OpenAPI URL from command flags or environment
func getOpenAPIURL(cmd *cobra.Command) string {
	if cmd.PersistentFlags().Lookup("openapi").Changed {
		return cmd.PersistentFlags().Lookup("openapi").Value.String()
	}
	if url := os.Getenv("QURL_OPENAPI"); url != "" {
		return url
	}
	return ""
}

func init() {
	rootCmd = &cobra.Command{
		Use:   "qurl [path]",
		Short: "Make HTTP requests with OpenAPI-powered help",
		Long: `qurl is a command-line HTTP client with OpenAPI integration.
It can display OpenAPI documentation and make HTTP requests with intelligent completion.`,
		Args:              cobra.MaximumNArgs(1),
		RunE:              runQurl,
		ValidArgsFunction: pathCompletion,
		SilenceUsage:      true,
	}

	// Global flags
	rootCmd.PersistentFlags().StringVar(&openAPIURL, "openapi", "", "OpenAPI specification URL")
	rootCmd.PersistentFlags().StringVarP(&httpMethod, "request", "X", "GET", "HTTP method (GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS)")
	rootCmd.PersistentFlags().BoolVar(&showDocs, "docs", false, "Show OpenAPI documentation for the specified endpoint")
	rootCmd.PersistentFlags().StringSliceVarP(&queryParams, "param", "p", []string{}, "Query parameters (can be used multiple times)")
	rootCmd.PersistentFlags().StringSliceVarP(&headers, "header", "H", []string{}, "Pass custom header(s) to server (can be used multiple times)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output (show request and response details)")
	rootCmd.PersistentFlags().BoolVarP(&includeHeaders, "include", "i", false, "Include response headers in output")
	rootCmd.PersistentFlags().StringVarP(&data, "data", "d", "", "HTTP POST/PUT/PATCH data")
	rootCmd.PersistentFlags().StringVar(&server, "server", "", "Server URL or index (overrides OpenAPI servers)")
	// Custom handling for --sig-v4 flag that can work with or without arguments
	rootCmd.PersistentFlags().BoolVar(&sigV4Enabled, "sig-v4", false, "Enable AWS SigV4 signing")
	rootCmd.PersistentFlags().StringVar(&sigV4Service, "sig-v4-service", "execute-api", "AWS service name for SigV4 signing")
	rootCmd.PersistentFlags().StringVar(&bearerToken, "bearer", "", "Bearer token for Authorization header")

	// Environment variable binding (QURL_ prefixed to avoid collisions)
	if url := os.Getenv("QURL_OPENAPI"); url != "" && openAPIURL == "" {
		openAPIURL = url
	}
	if srv := os.Getenv("QURL_SERVER"); srv != "" && server == "" {
		server = srv
	}

	// Register completion functions
	rootCmd.RegisterFlagCompletionFunc("request", methodCompletion)
	rootCmd.RegisterFlagCompletionFunc("X", methodCompletion)
	rootCmd.RegisterFlagCompletionFunc("param", paramCompletion)
	rootCmd.RegisterFlagCompletionFunc("server", serverCompletion)

	// Generate completion commands
	rootCmd.AddCommand(generateCompletionCmd())
}

func runQurl(cmd *cobra.Command, args []string) error {
	// Get OpenAPI URL from flag or environment
	openAPIURL = getOpenAPIURL(cmd)

	path := ""
	if len(args) > 0 {
		path = args[0]
	}

	// Normalize HTTP method
	httpMethod = strings.ToUpper(httpMethod)

	// If docs flag is set, show OpenAPI documentation
	if showDocs {
		if openAPIURL == "" {
			return fmt.Errorf("OpenAPI URL is required for docs (use --openapi flag or QURL_OPENAPI env var)")
		}
		// For docs, show all methods unless explicitly specified with -X
		docsMethod := httpMethod
		if httpMethod == "GET" && !cmd.Flags().Changed("request") {
			docsMethod = "ANY"
		}
		return showOpenAPIDocumentation(path, docsMethod)
	}

	// If no path specified and no docs flag, show usage
	if path == "" {
		return cmd.Help()
	}

	// Execute HTTP request
	return executeHTTPRequest(path, httpMethod)
}

// createViewer creates a viewer instance with the specified OpenAPI URL
// If url is empty, uses the global openAPIURL
func createViewer(url string) (*openapi.Viewer, error) {
	if url == "" {
		url = openAPIURL
	}
	httpClient, err := qurlhttp.NewClient()
	if err != nil {
		return nil, fmt.Errorf("error creating HTTP client: %w", err)
	}
	return openapi.NewViewer(httpClient, url), nil
}

func showOpenAPIDocumentation(path, method string) error {
	viewer, err := createViewer("")
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// If no path specified, show all endpoints
	if path == "" {
		path = "*"
	}

	// If no method specified, show all methods
	if method == "" {
		method = "ANY"
	}

	output, err := viewer.View(ctx, path, method)
	if err != nil {
		return fmt.Errorf("error viewing OpenAPI spec: %w", err)
	}

	fmt.Println(output)
	return nil
}

// resolveTargetURL determines the final URL for the request based on the path and configuration
func resolveTargetURL(path string, viewer *openapi.Viewer) (string, error) {
	// Parse the path to check if it's an absolute URL
	parsedURL, err := url.Parse(path)
	if err != nil {
		return "", fmt.Errorf("invalid URL/path: %w", err)
	}

	// Check if we have a full URL (with scheme and host)
	if parsedURL.Scheme != "" && parsedURL.Host != "" {
		// Full URL provided - use it directly
		return path, nil
	}

	// Path only - need to get base URL
	var baseURL string

	// Priority 1: --server flag provided
	if server != "" {
		baseURL, err = resolveServerURL(server, viewer)
		if err != nil {
			return "", fmt.Errorf("failed to resolve server URL: %w", err)
		}
	} else if openAPIURL != "" {
		// Priority 2: Get base URL from OpenAPI spec
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		baseURL, err = viewer.BaseURL(ctx)
		if err != nil {
			return "", fmt.Errorf("failed to get base URL from OpenAPI spec: %w", err)
		}
	} else {
		// Priority 3: Error - no way to construct URL
		return "", fmt.Errorf("server URL required when using relative paths (use --server flag, --openapi flag, or full URL)")
	}

	// Combine base URL with path
	base, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("invalid base URL from OpenAPI: %w", err)
	}

	// Handle relative path
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	// Properly append path to base URL
	// If base already has a path, we need to combine them
	if base.Path != "" && base.Path != "/" {
		base.Path = strings.TrimSuffix(base.Path, "/") + path
	} else {
		base.Path = path
	}
	return base.String(), nil
}

// executeAndHandleResponse executes the HTTP request and handles the response output
func executeAndHandleResponse(client *qurlhttp.Client, req *http.Request, method, targetURL string) error {
	// Show request details if verbose
	if verbose {
		fmt.Fprintf(os.Stderr, "> %s %s\n", method, targetURL)
		fmt.Fprintf(os.Stderr, "> Host: %s\n", req.URL.Host)
		// Show all headers
		for name, values := range req.Header {
			for _, value := range values {
				fmt.Fprintf(os.Stderr, "> %s: %s\n", name, value)
			}
		}
		fmt.Fprintf(os.Stderr, ">\n")
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response: %w", err)
	}

	// Print response details based on flags
	if verbose {
		fmt.Fprintf(os.Stderr, "< HTTP/%d.%d %s\n", resp.ProtoMajor, resp.ProtoMinor, resp.Status)
		for key, values := range resp.Header {
			for _, value := range values {
				fmt.Fprintf(os.Stderr, "< %s: %s\n", key, value)
			}
		}
		fmt.Fprintf(os.Stderr, "<\n")
	} else if includeHeaders {
		// Print status line and headers to stdout (like curl -i)
		fmt.Printf("HTTP/%d.%d %s\n", resp.ProtoMajor, resp.ProtoMinor, resp.Status)
		for key, values := range resp.Header {
			for _, value := range values {
				fmt.Printf("%s: %s\n", key, value)
			}
		}
		fmt.Println() // Empty line between headers and body
	}

	// Always print response body
	fmt.Print(string(body))

	return nil
}

// buildHTTPRequest creates and configures the HTTP request with all headers and body
func buildHTTPRequest(method, targetURL, path string, viewer *openapi.Viewer) (*http.Request, error) {
	// Create the request with body if data is provided
	var requestBody io.Reader
	if data != "" {
		requestBody = strings.NewReader(data)
	}

	req, err := http.NewRequest(method, targetURL, requestBody)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Set standard headers
	req.Header.Set("User-Agent", "qurl")

	// Set headers based on OpenAPI spec if available
	if viewer != nil && path != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := viewer.SetHeaders(ctx, req, path, method); err != nil {
			// Log warning but continue - headers are not critical
			if verbose {
				fmt.Fprintf(os.Stderr, "Warning: could not set headers from OpenAPI: %v\n", err)
			}
		}
	}

	// Apply authentication first (can be overridden by custom headers)
	if err := applyAuthentication(req, targetURL); err != nil {
		return nil, fmt.Errorf("applying authentication: %w", err)
	}

	// Set custom headers from -H flags (these override any headers set above)
	for _, header := range headers {
		// Split on first : to get header name and value
		parts := strings.SplitN(header, ":", 2)
		if len(parts) == 2 {
			headerName := strings.TrimSpace(parts[0])
			headerValue := strings.TrimSpace(parts[1])
			req.Header.Set(headerName, headerValue)
		} else if len(parts) == 1 && parts[0] != "" {
			// Handle header without value (just name)
			headerName := strings.TrimSpace(parts[0])
			req.Header.Set(headerName, "")
		}
	}

	// Set Content-Type header if data is provided and no custom Content-Type was set
	if data != "" && req.Header.Get("Content-Type") == "" {
		// Try to detect content type
		if strings.HasPrefix(strings.TrimSpace(data), "{") && strings.HasSuffix(strings.TrimSpace(data), "}") {
			req.Header.Set("Content-Type", "application/json")
		} else {
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		}
	}

	return req, nil
}

// applyQueryParameters adds query parameters from flags to the target URL
func applyQueryParameters(targetURL string) (string, error) {
	if len(queryParams) == 0 {
		return targetURL, nil
	}

	parsedTarget, err := url.Parse(targetURL)
	if err != nil {
		return "", fmt.Errorf("error parsing target URL: %w", err)
	}

	query := parsedTarget.Query()
	for _, param := range queryParams {
		// Split on first = to get key and value
		parts := strings.SplitN(param, "=", 2)
		if len(parts) == 2 {
			query.Add(parts[0], parts[1])
		} else if len(parts) == 1 && parts[0] != "" {
			// Handle param without value (just key)
			query.Add(parts[0], "")
		}
	}
	parsedTarget.RawQuery = query.Encode()
	return parsedTarget.String(), nil
}

func executeHTTPRequest(path, method string) error {
	// Create HTTP client with Lambda support
	client, err := qurlhttp.NewClient()
	if err != nil {
		return fmt.Errorf("error creating HTTP client: %w", err)
	}

	// Create viewer for OpenAPI operations if spec is available
	var viewer *openapi.Viewer
	if openAPIURL != "" {
		viewer, err = createViewer("")
		if err != nil {
			return err
		}
	}

	// Resolve the target URL
	targetURL, err := resolveTargetURL(path, viewer)
	if err != nil {
		return err
	}

	// Add query parameters
	targetURL, err = applyQueryParameters(targetURL)
	if err != nil {
		return err
	}

	// Build the HTTP request
	req, err := buildHTTPRequest(method, targetURL, path, viewer)
	if err != nil {
		return err
	}

	// Execute and handle the response
	return executeAndHandleResponse(client, req, method, targetURL)
}

// resolveServerURL resolves the server URL based on the --server flag value
// It can be either a full URL or a numeric index to select from OpenAPI servers
func resolveServerURL(serverFlag string, viewer *openapi.Viewer) (string, error) {
	// Check if it's a numeric index (0, 1, 2, etc.)
	if len(serverFlag) == 1 && serverFlag[0] >= '0' && serverFlag[0] <= '9' {
		// Parse as server index
		index := int(serverFlag[0] - '0')

		if viewer == nil {
			return "", fmt.Errorf("server index '%s' requires OpenAPI specification", serverFlag)
		}

		servers, err := viewer.GetServers()
		if err != nil {
			return "", fmt.Errorf("failed to get servers from OpenAPI spec: %w", err)
		}

		if index >= len(servers) {
			return "", fmt.Errorf("server index %d out of range (available: 0-%d)", index, len(servers)-1)
		}

		serverURL := servers[index].URL
		if serverURL == "" {
			return "", fmt.Errorf("server at index %d has empty URL", index)
		}

		// Handle relative URLs
		if !strings.HasPrefix(serverURL, "http://") && !strings.HasPrefix(serverURL, "https://") {
			if viewer == nil {
				return "", fmt.Errorf("relative server URL requires OpenAPI specification")
			}
			// Use the viewer's logic to resolve relative URLs
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			return viewer.BaseURL(ctx)
		}

		return serverURL, nil
	}

	// Treat as full URL
	parsedURL, err := url.Parse(serverFlag)
	if err != nil {
		return "", fmt.Errorf("invalid server URL: %w", err)
	}

	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return "", fmt.Errorf("server URL must be complete (e.g., https://example.com)")
	}

	return serverFlag, nil
}

// Completion functions

func methodCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}
	return methods, cobra.ShellCompDirectiveNoFileComp
}

func serverCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	var suggestions []string

	// Add common URL patterns
	suggestions = append(suggestions, "http://localhost:8080", "https://localhost:8443", "http://localhost:3000")

	// Get OpenAPI URL to provide server suggestions
	openapiURL := cmd.Flag("openapi").Value.String()
	if openapiURL == "" {
		if url := os.Getenv("QURL_OPENAPI"); url != "" {
			openapiURL = url
		}
	}

	if openapiURL != "" {
		// Try to get servers from OpenAPI spec
		httpClient, err := qurlhttp.NewClient()
		if err == nil {
			viewer := openapi.NewViewer(httpClient, openapiURL)

			if servers, err := viewer.GetServers(); err == nil {
				// Add server indices and URLs
				for i, server := range servers {
					suggestions = append(suggestions, fmt.Sprintf("%d", i))
					if server.URL != "" {
						suggestions = append(suggestions, server.URL)
					}
				}
			}
		}
	}

	return suggestions, cobra.ShellCompDirectiveNoFileComp
}

func paramCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// Get OpenAPI URL
	openapiURL := getOpenAPIURL(cmd)

	if openapiURL == "" {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// Get the path from args (if provided)
	path := ""
	if len(args) > 0 {
		path = args[0]
	}

	// If no path yet, can't suggest parameters
	if path == "" {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// Get HTTP method
	method := cmd.Flag("request").Value.String()
	if method == "" {
		method = "GET"
	}
	method = strings.ToUpper(method)

	// Create viewer
	viewer, err := createViewer(openapiURL)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get parameter completions
	params, err := viewer.ParamCompletions(ctx, path, method)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	// Format completions as key=value suggestions
	var completions []string
	for _, param := range params {
		// If user already typed part of param=value, complete it
		if strings.Contains(toComplete, "=") {
			// User is typing the value part, don't suggest
			return nil, cobra.ShellCompDirectiveNoSpace
		}
		// Suggest param= format
		completions = append(completions, param+"=")
	}

	return completions, cobra.ShellCompDirectiveNoSpace
}

func pathCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// Only complete if we haven't specified a path argument yet
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// Get OpenAPI URL
	openapiURL := getOpenAPIURL(cmd)

	if openapiURL == "" {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// Get HTTP method for filtering
	method := cmd.Flag("request").Value.String()
	if method == "" {
		method = "GET"
	}
	method = strings.ToUpper(method)

	// Create viewer
	viewer, err := createViewer(openapiURL)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get path completions
	paths, err := viewer.PathCompletions(ctx, toComplete, method)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	return paths, cobra.ShellCompDirectiveNoFileComp
}

func generateCompletionCmd() *cobra.Command {
	completionCmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate completion script",
		Long: `To load completions:

Bash:

  # Load for current session:
  $ source <(qurl completion bash)

  # Load for all sessions (add to ~/.bashrc):
  $ echo 'source <(qurl completion bash)' >> ~/.bashrc

Zsh:

  # Load for current session:
  $ source <(qurl completion zsh)

  # Load for all sessions (add to ~/.zshrc):
  $ echo 'source <(qurl completion zsh)' >> ~/.zshrc

Fish:

  # Load for current session:
  $ qurl completion fish | source

  # Load for all sessions:
  $ qurl completion fish > ~/.config/fish/completions/qurl.fish

PowerShell:

  # Load for current session:
  PS> qurl completion powershell | Out-String | Invoke-Expression

  # Load for all sessions (add to PowerShell profile):
  PS> Add-Content $PROFILE 'qurl completion powershell | Out-String | Invoke-Expression'
`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		Run: func(cmd *cobra.Command, args []string) {
			switch args[0] {
			case "bash":
				cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				cmd.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
			}
		},
	}

	return completionCmd
}

// applyAuthentication applies authentication to the request based on flags
func applyAuthentication(req *http.Request, targetURL string) error {
	// Apply bearer token if provided (can be overridden by custom headers later)
	if bearerToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", bearerToken))
	}

	// Check if this is a lambda:// URL - skip SigV4 for direct invocation
	if strings.HasPrefix(targetURL, "lambda://") {
		// Lambda invocations use AWS SDK authentication internally, but bearer tokens can still be applied
		return nil
	}

	// Apply AWS SigV4 signing if requested
	if sigV4Enabled {
		service := sigV4Service

		// Load AWS config from default credential chain
		ctx := context.Background()
		cfg, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			return fmt.Errorf("loading AWS config: %w", err)
		}

		// Extract region from the URL or use default
		region := cfg.Region

		// Retrieve credentials
		creds, err := cfg.Credentials.Retrieve(ctx)
		if err != nil {
			return fmt.Errorf("retrieving AWS credentials: %w", err)
		}

		// Create signer
		signer := v4.NewSigner()

		// Calculate payload hash for the signature
		var payloadHash string
		if req.Body != nil {
			// Read body to calculate hash
			bodyBytes, err := io.ReadAll(req.Body)
			if err != nil {
				return fmt.Errorf("reading request body for signing: %w", err)
			}

			// Calculate SHA256 hash
			hash := sha256.Sum256(bodyBytes)
			payloadHash = fmt.Sprintf("%x", hash)

			// Restore body for actual request
			req.Body = io.NopCloser(strings.NewReader(string(bodyBytes)))
		} else {
			// Empty body hash
			hash := sha256.Sum256([]byte{})
			payloadHash = fmt.Sprintf("%x", hash)
		}

		// Sign the request
		err = signer.SignHTTP(ctx, creds, req, payloadHash, service, region, time.Now())
		if err != nil {
			return fmt.Errorf("signing request with SigV4: %w", err)
		}

		if verbose {
			fmt.Fprintf(os.Stderr, "Applied AWS SigV4 signature for service: %s, region: %s\n", service, region)
		}
	}

	return nil
}
