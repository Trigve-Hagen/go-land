package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestIndex(t *testing.T) {
	req, err := http.NewRequest(
		http.MethodPost,
		"http://localhost:3000/index",
		strings.NewReader("nemail=trigve.hagen@gmail.com"),
	)
	if err != nil {
		t.Fatalf("could not create the request: %v", err)
	}

	rec := httptest.NewRecorder()
	index(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200; got %d", rec.Code)
	}
}
