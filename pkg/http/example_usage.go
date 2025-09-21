package http

// Example usage of the HTTP client with Lambda support
//
// Standard HTTP requests:
//   client, _ := NewClient()
//   resp, _ := client.Get("https://api.example.com/users")
//
// Lambda invocations:
//   client, _ := NewClient()
//   resp, _ := client.Get("lambda://my-function/users")
//
// The Lambda URL format:
//   lambda://<function-name>/<path>?<query-params>
//
// Examples:
//   lambda://user-service/users/123
//   lambda://auth-service/login
//   lambda://data-processor/process?format=json&limit=100
//
// The client automatically:
// - Detects lambda:// URLs
// - Converts HTTP requests to API Gateway v2 proxy events
// - Invokes the Lambda function
// - Converts the Lambda response back to an HTTP response
//
// Lambda functions should handle API Gateway v2 HTTP proxy events
// and return responses in the expected format:
//
//   {
//     "statusCode": 200,
//     "headers": {
//       "Content-Type": "application/json"
//     },
//     "body": "{\"message\":\"success\"}",
//     "isBase64Encoded": false
//   }
