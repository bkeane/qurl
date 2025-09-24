// Package testutil provides shared testing utilities and fixtures
// This eliminates the massive duplication across 15+ test files
package testutil

// Common OpenAPI spec fixtures used across multiple test files
const (
	// BasicAPISpec provides a minimal valid OpenAPI 3.0 spec
	BasicAPISpec = `{
		"openapi": "3.0.0",
		"info": {
			"title": "Test API",
			"version": "1.0.0"
		},
		"paths": {
			"/users": {
				"get": {
					"responses": {
						"200": {
							"description": "Success",
							"content": {
								"application/json": {
									"schema": {"type": "array", "items": {"type": "object"}}
								}
							}
						}
					}
				},
				"post": {
					"requestBody": {
						"content": {
							"application/json": {
								"schema": {"type": "object"}
							}
						}
					},
					"responses": {
						"201": {
							"description": "Created",
							"content": {
								"application/json": {
									"schema": {"type": "object"}
								}
							}
						}
					}
				}
			},
			"/users/{id}": {
				"get": {
					"parameters": [
						{
							"name": "id",
							"in": "path",
							"required": true,
							"schema": {"type": "string"}
						}
					],
					"responses": {
						"200": {
							"description": "Success",
							"content": {
								"application/json": {
									"schema": {"type": "object"}
								}
							}
						}
					}
				}
			}
		}
	}`

	// MultiContentTypeSpec provides specs with multiple content types for Accept header testing
	MultiContentTypeSpec = `{
		"openapi": "3.0.0",
		"info": {
			"title": "Multi Content API",
			"version": "1.0.0"
		},
		"paths": {
			"/data": {
				"get": {
					"responses": {
						"200": {
							"description": "Success",
							"content": {
								"application/json": {
									"schema": {"type": "object"}
								},
								"application/xml": {
									"schema": {"type": "object"}
								},
								"text/plain": {
									"schema": {"type": "string"}
								}
							}
						}
					}
				}
			}
		}
	}`

	// ServerAPISpec provides specs with multiple servers for testing server resolution
	ServerAPISpec = `{
		"openapi": "3.0.0",
		"info": {
			"title": "Server Test API",
			"version": "1.0.0"
		},
		"servers": [
			{"url": "https://api.example.com"},
			{"url": "https://staging.example.com"},
			{"url": "https://dev.example.com/api/v1"}
		],
		"paths": {
			"/test": {
				"get": {
					"responses": {
						"200": {
							"description": "Success",
							"content": {
								"application/json": {
									"schema": {"type": "object"}
								}
							}
						}
					}
				}
			}
		}
	}`

	// ParameterAPISpec provides comprehensive parameter testing scenarios
	ParameterAPISpec = `{
		"openapi": "3.0.0",
		"info": {
			"title": "Parameter Test API",
			"version": "1.0.0"
		},
		"paths": {
			"/search": {
				"get": {
					"parameters": [
						{
							"name": "query",
							"in": "query",
							"required": true,
							"schema": {"type": "string"}
						},
						{
							"name": "limit",
							"in": "query",
							"required": false,
							"schema": {"type": "integer", "default": 10}
						},
						{
							"name": "offset",
							"in": "query",
							"required": false,
							"schema": {"type": "integer", "default": 0}
						}
					],
					"responses": {
						"200": {
							"description": "Search results",
							"content": {
								"application/json": {
									"schema": {"type": "array", "items": {"type": "object"}}
								}
							}
						}
					}
				}
			},
			"/items/{category}": {
				"get": {
					"parameters": [
						{
							"name": "category",
							"in": "path",
							"required": true,
							"schema": {"type": "string"}
						},
						{
							"name": "filter",
							"in": "query",
							"schema": {"type": "string"}
						}
					],
					"responses": {
						"200": {
							"description": "Items in category",
							"content": {
								"application/json": {
									"schema": {"type": "array", "items": {"type": "object"}}
								}
							}
						}
					}
				}
			}
		}
	}`

	// EmptyAPISpec provides minimal spec for error testing
	EmptyAPISpec = `{
		"openapi": "3.0.0",
		"info": {
			"title": "Empty API",
			"version": "1.0.0"
		},
		"paths": {}
	}`
)