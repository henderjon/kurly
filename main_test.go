package main

import (
	"net/http"
	"strings"
	"testing"
)

func TestSetHeaders(t *testing.T) {
	header := []string{"User-Agent: Curly/1.0"}
	req, err := http.NewRequest("GET", "http://url.com/", nil)
	if err != nil {
		panic(err)
	}

	setHeaders(req, header)

	if len(req.Header) > 0 {
		for _, v := range req.Header {
			userAgentValue := strings.Join(v, "")
			if userAgentValue != "Curly/1.0" {
				t.Errorf("Expected Curly/1.0, but got %g", userAgentValue)
			}
		}
	} else {
		t.Error("setHeaders() set no header")
	}
}
