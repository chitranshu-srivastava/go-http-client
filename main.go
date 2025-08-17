package main

import (
	"bufio"
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Config struct {
	Method   string
	URL      string
	Headers  []string
	Query    []string
	Data     string
	Form     []string
	Timeout  time.Duration
}

type HeaderList []string

func (h *HeaderList) String() string {
	return strings.Join(*h, ", ")
}

func (h *HeaderList) Set(value string) error {
	*h = append(*h, value)
	return nil
}

type QueryList []string

func (q *QueryList) String() string {
	return strings.Join(*q, ", ")
}

func (q *QueryList) Set(value string) error {
	*q = append(*q, value)
	return nil
}

type FormList []string

func (f *FormList) String() string {
	return strings.Join(*f, ", ")
}

func (f *FormList) Set(value string) error {
	*f = append(*f, value)
	return nil
}

func main() {
	var config Config
	var headers HeaderList
	var queries QueryList
	var forms FormList

	flag.StringVar(&config.Method, "X", "GET", "HTTP method")
	flag.StringVar(&config.Method, "method", "GET", "HTTP method")
	flag.Var(&headers, "H", "Header in 'Key: Value' format")
	flag.Var(&headers, "header", "Header in 'Key: Value' format")
	flag.Var(&queries, "q", "Query parameter in 'key=value' format")
	flag.Var(&queries, "query", "Query parameter in 'key=value' format")
	flag.StringVar(&config.Data, "d", "", "Request data (string, @filename, or - for stdin)")
	flag.StringVar(&config.Data, "data", "", "Request data (string, @filename, or - for stdin)")
	flag.Var(&forms, "f", "Form data in 'key=value' or 'key=@filename' format")
	flag.Var(&forms, "form", "Form data in 'key=value' or 'key=@filename' format")
	flag.DurationVar(&config.Timeout, "t", 30*time.Second, "Request timeout")
	flag.DurationVar(&config.Timeout, "timeout", 30*time.Second, "Request timeout")

	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] URL\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}

	config.URL = flag.Arg(0)
	config.Headers = headers
	config.Query = queries
	config.Form = forms

	if err := makeRequest(config); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func makeRequest(config Config) error {
	parsedURL, err := url.Parse(config.URL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	var body io.Reader
	var contentType string

	if len(config.Form) > 0 {
		body, contentType, err = buildFormData(config.Form)
		if err != nil {
			return fmt.Errorf("failed to build form data: %w", err)
		}
	} else if config.Data != "" {
		body, err = buildRequestBody(config.Data)
		if err != nil {
			return fmt.Errorf("failed to build request body: %w", err)
		}
	}

	req, err := http.NewRequest(config.Method, parsedURL.String(), body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	addHeaders(req, config.Headers)
	addQueryParams(req, config.Query)

	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()
	req = req.WithContext(ctx)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	fmt.Printf("%s %s\n", resp.Proto, resp.Status)
	for key, values := range resp.Header {
		for _, value := range values {
			fmt.Printf("%s: %s\n", key, value)
		}
	}
	fmt.Println()

	_, err = io.Copy(os.Stdout, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	return nil
}

func buildRequestBody(data string) (io.Reader, error) {
	if data == "" {
		return nil, nil
	}

	if data == "-" {
		scanner := bufio.NewScanner(os.Stdin)
		var lines []string
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("failed to read from stdin: %w", err)
		}
		return strings.NewReader(strings.Join(lines, "\n")), nil
	}

	if strings.HasPrefix(data, "@") {
		filename := data[1:]
		file, err := os.Open(filename)
		if err != nil {
			return nil, fmt.Errorf("failed to open file %s: %w", filename, err)
		}
		defer file.Close()

		content, err := io.ReadAll(file)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", filename, err)
		}
		return bytes.NewReader(content), nil
	}

	return strings.NewReader(data), nil
}

func buildFormData(forms []string) (io.Reader, string, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	for _, form := range forms {
		parts := strings.SplitN(form, "=", 2)
		if len(parts) != 2 {
			return nil, "", fmt.Errorf("invalid form data format: %s", form)
		}

		key := parts[0]
		value := parts[1]

		if strings.HasPrefix(value, "@") {
			filename := value[1:]
			file, err := os.Open(filename)
			if err != nil {
				return nil, "", fmt.Errorf("failed to open file %s: %w", filename, err)
			}
			defer file.Close()

			part, err := writer.CreateFormFile(key, filepath.Base(filename))
			if err != nil {
				return nil, "", fmt.Errorf("failed to create form file: %w", err)
			}

			_, err = io.Copy(part, file)
			if err != nil {
				return nil, "", fmt.Errorf("failed to copy file content: %w", err)
			}
		} else {
			err := writer.WriteField(key, value)
			if err != nil {
				return nil, "", fmt.Errorf("failed to write form field: %w", err)
			}
		}
	}

	err := writer.Close()
	if err != nil {
		return nil, "", fmt.Errorf("failed to close multipart writer: %w", err)
	}

	return &buf, writer.FormDataContentType(), nil
}

func addHeaders(req *http.Request, headers []string) {
	for _, header := range headers {
		parts := strings.SplitN(header, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			req.Header.Set(key, value)
		}
	}
}

func addQueryParams(req *http.Request, queries []string) {
	q := req.URL.Query()
	for _, query := range queries {
		parts := strings.SplitN(query, "=", 2)
		if len(parts) == 2 {
			q.Add(parts[0], parts[1])
		}
	}
	req.URL.RawQuery = q.Encode()
}