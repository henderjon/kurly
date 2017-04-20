package main

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestSetHeaders(t *testing.T) {
	var body io.Reader
	header := []string{"User-Agent: Curly/1.0"}
	req, err := http.NewRequest("GET", "http://url.com/", body)
	if err != nil {
		panic(err)
	}

	setHeaders(req, header)

	for _, v := range req.Header {
		userAgentValue := strings.Join(v, "")
		if userAgentValue != "Curly/1.0" {
			t.Errorf("Expected Curly/1.0, but got %g", userAgentValue)
		}
	}
}
