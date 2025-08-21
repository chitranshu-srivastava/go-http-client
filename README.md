# Go Http Client

* Clone the project
* *go mod download* to install dependencies
* *go build* to build the binary
* *./http-client [OPTIONS] URL* to use it


## Usage Examples:

``` ./http-client --help ```

## Simple GET request

```./http-client https://api.github.com/users/octocat```

## POST with JSON data

```./http-client -X POST -H "Content-Type: application/json" -d '{"name":"test"}' https://httpbin.org/post```

## GET with query parameters and headers

```./http-client -q "page=1" -q "limit=10" -H "Authorization: Bearer token" https://api.example.com/data```

## Upload file via form

```./http-client -X POST -f "file=@document.pdf" -f "description=My file" https://httpbin.org/post```

## Read data from stdin

```echo "test data" | ./http-client -X POST -d - https://httpbin.org/post```

## Custom timeout

```./http-client -t 5s https://slow-api.example.com```

## Basic Authentication

```./http-client -u username -p password https://api.example.com```

## Bearer Token

```./http-client -b "your-token-here" https://api.example.com```

## OAuth2 Client Credentials

```./http-client --client-id "client123" --client-secret "secret456" --token-url "https://auth.example.com/token" --scope "read" --scope "write" https://api.example.com```

## Custom Authentication Header

```./http-client --auth-header "X-API-Key" --auth-value "your-api-key" https://api.example.com```

## Rate Limiting

The HTTP client supports rate limiting using the Token Bucket algorithm to control request frequency. This is useful for respecting API rate limits and preventing server overload.

### Rate Limiting Options

- `--rate` or `-r`: Set the rate limit in format `requests/duration`

### Rate Limit Format

The rate limit string accepts various formats:

- `10/s` - 10 requests per second
- `100/30s` - 100 requests per 30 seconds  
- `50/m` - 50 requests per minute
- `1000/h` - 1000 requests per hour
- `5/2m` - 5 requests per 2 minutes

### Rate Limiting Examples

```bash
# Limit to 10 requests per second
./http-client --rate "10/s" https://api.example.com

# Limit to 100 requests per 30 seconds
./http-client -r "100/30s" https://api.example.com

# Combined with authentication
./http-client -b "token" --rate "50/m" https://api.example.com
```

### How Rate Limiting Works

- Uses the Token Bucket algorithm from `golang.org/x/time/rate`
- Allows burst traffic up to the specified limit
- Automatically waits when rate limit is exceeded
- Integrates with request timeout settings
- Thread-safe for concurrent usage

### Rate Limiting Behavior

- **Burst Capacity**: The burst size equals the number of requests in the rate specification
- **Token Replenishment**: Tokens are added at the specified rate
- **Blocking**: When rate limit is exceeded, the client waits for available tokens
- **Timeout Integration**: Rate limiting waits respect the overall request timeout
