package response

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Formatter interface {
	Format(resp *http.Response) ([]byte, error)
}

type PrettyFormatter struct{}

func NewPrettyFormatter() *PrettyFormatter {
	return &PrettyFormatter{}
}

func (pf *PrettyFormatter) Format(resp *http.Response) ([]byte, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	contentType := resp.Header.Get("Content-Type")
	
	if strings.Contains(contentType, "application/json") || strings.Contains(contentType, "text/json") {
		return pf.formatJSON(body)
	}
	
	if strings.Contains(contentType, "application/xml") || strings.Contains(contentType, "text/xml") {
		return pf.formatXML(body)
	}
	
	return body, nil
}

func (pf *PrettyFormatter) formatJSON(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return data, nil
	}
	
	var obj interface{}
	if err := json.Unmarshal(data, &obj); err != nil {
		return data, nil
	}
	
	pretty, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return data, nil
	}
	
	return pretty, nil
}

func (pf *PrettyFormatter) formatXML(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return data, nil
	}
	
	var buf bytes.Buffer
	var formatted bytes.Buffer
	
	if err := xml.Unmarshal(data, &buf); err != nil {
		return data, nil
	}
	
	encoder := xml.NewEncoder(&formatted)
	encoder.Indent("", "  ")
	
	decoder := xml.NewDecoder(bytes.NewReader(data))
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return data, nil
		}
		
		if err := encoder.EncodeToken(token); err != nil {
			return data, nil
		}
	}
	
	if err := encoder.Flush(); err != nil {
		return data, nil
	}
	
	return formatted.Bytes(), nil
}

type RawFormatter struct{}

func NewRawFormatter() *RawFormatter {
	return &RawFormatter{}
}

func (rf *RawFormatter) Format(resp *http.Response) ([]byte, error) {
	return io.ReadAll(resp.Body)
}