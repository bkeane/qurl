# qurl Test Specifications

This directory contains the comprehensive test suite for qurl, organized with clear separation of concerns between implementation verification and feature documentation.

## Test Philosophy

### Two Distinct Testing Approaches

#### 1. **httpbin_spec.sh** - Deep Request Layer Verification
**Purpose:** Exhaustive verification of HTTP request generation correctness

- **Scope:** In-the-weeds validation of every aspect of request construction
- **Strategy:** Uses httpbin.org's echo endpoints to inspect exact request structure
- **Coverage:** Both OpenAPI-enabled (using httpbin's spec) and vanilla HTTP modes
- **Focus Areas:**
  - HTTP method generation and normalization
  - Header construction and management
  - Query parameter encoding and edge cases
  - URL/path handling
  - Request body formatting (when implemented)
  - OpenAPI-driven enrichment (Accept headers, base URL extraction)
  - Error conditions and edge cases
  - curl compatibility features

**Test Structure:**
```
With OpenAPI specification
├── OpenAPI-driven request enrichment
└── Header generation with OpenAPI context

Without OpenAPI (vanilla HTTP)
├── Core request construction
├── Verbose output verification
├── Response handling
├── Error conditions
└── curl compatibility
```

#### 2. **petstore_spec.sh** - User Feature Documentation
**Purpose:** Human-readable demonstration of qurl's capabilities

- **Scope:** Wide coverage of user-facing features
- **Strategy:** Real-world API workflows using Swagger Petstore
- **Coverage:** Complete feature set from a user's perspective
- **Focus Areas:**
  - API exploration with documentation
  - Making API calls with OpenAPI awareness
  - Different HTTP methods
  - Debugging with verbose mode
  - Response formatting options
  - Custom headers and overrides
  - Environment configuration
  - Practical workflows

**Test Structure:**
```
Feature-based organization
├── Exploring APIs with documentation
├── Making API calls with OpenAPI awareness
├── Different HTTP methods
├── Verbose mode for debugging
├── Response formatting options
├── Error handling
├── Custom headers
├── Shell completion
└── Practical API workflows
```

## Key Distinction

- **httpbin tests:** "Does qurl correctly construct HTTP requests?" (implementation correctness)
- **petstore tests:** "What can a user do with qurl?" (feature demonstration)

## Running Tests

### Prerequisites
Install ShellSpec:
```bash
curl -fsSL https://git.io/shellspec | sh
```

### Run All Tests
```bash
./run_e2e_tests.sh
```

### Run Specific Test Suites
```bash
# Deep request verification
shellspec spec/httpbin_spec.sh

# User feature documentation
shellspec spec/petstore_spec.sh
```

### Run with Different Output Formats
```bash
# Detailed documentation format (best for understanding features)
shellspec --format documentation spec/petstore_spec.sh

# TAP format for CI integration
shellspec --format tap

# Progress dots for quick feedback
shellspec --format progress
```

## Test Coverage Matrix

### httpbin_spec.sh Coverage
| Category | With OpenAPI | Without OpenAPI |
|----------|--------------|-----------------|
| HTTP Methods | ✓ | ✓ |
| Headers | ✓ | ✓ |
| Query Parameters | ✓ | ✓ |
| Request Body | Planned | Planned |
| Accept Header | ✓ (auto) | - |
| Base URL | ✓ (from spec) | ✓ (explicit) |
| Verbose Output | ✓ | ✓ |
| Error Handling | ✓ | ✓ |
| Edge Cases | ✓ | ✓ |

### petstore_spec.sh Feature Coverage
| Feature | Status |
|---------|---------|
| API Documentation Viewing | ✓ |
| OpenAPI-aware Requests | ✓ |
| HTTP Methods (GET, HEAD, OPTIONS) | ✓ |
| POST/PUT with Body | Planned |
| Verbose Debugging | ✓ |
| Response Formatting | ✓ |
| Custom Headers | ✓ |
| Environment Configuration | ✓ |
| Shell Completions | ✓ (generation) |
| Error Handling | ✓ |

## Test Dependencies

- **httpbin.org** - For request inspection and verification
  - OpenAPI spec: https://httpbin.org/spec.json
  - Echo endpoints for header/parameter verification

- **Swagger Petstore API** - For real-world API interaction
  - OpenAPI spec: https://petstore3.swagger.io/api/v3/openapi.json
  - Live API endpoints for feature demonstration

## Writing New Tests

### For httpbin_spec.sh
Add tests when:
- Implementing new HTTP client features
- Fixing bugs in request generation
- Adding edge case handling
- Ensuring curl compatibility

Focus on:
- Precise verification using httpbin's echo
- Testing both with and without OpenAPI
- Edge cases and error conditions

### For petstore_spec.sh
Add tests when:
- Adding new user-facing features
- Demonstrating workflows
- Documenting capabilities

Focus on:
- Clear, readable test descriptions
- Real-world use cases
- Feature demonstration over implementation details

## Notes

- Tests require internet connectivity
- httpbin tests verify implementation correctness
- petstore tests serve as living documentation
- Both test suites complement each other for complete coverage