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
