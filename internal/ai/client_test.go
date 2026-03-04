package ai

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(req *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func newMockHTTPClient(fn roundTripFunc) *http.Client {
	return &http.Client{Transport: fn}
}

func TestSidecarClientHealth(t *testing.T) {
	client := NewSidecarClient("http://sidecar.local", 0)
	client.httpClient = newMockHTTPClient(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodGet {
			t.Fatalf("unexpected method: %s", req.Method)
		}
		if req.URL.String() != "http://sidecar.local/health" {
			t.Fatalf("unexpected URL: %s", req.URL.String())
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"status":"ok"}`)),
			Header:     make(http.Header),
		}, nil
	})

	health, err := client.Health(context.Background())
	if err != nil {
		t.Fatalf("Health failed: %v", err)
	}
	if health.Status != "ok" {
		t.Fatalf("unexpected health status: %q", health.Status)
	}
}

func TestSidecarClientRunCommand(t *testing.T) {
	client := NewSidecarClient("http://sidecar.local", 0)
	client.httpClient = newMockHTTPClient(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", req.Method)
		}
		if req.URL.String() != "http://sidecar.local/v1/ai/command" {
			t.Fatalf("unexpected URL: %s", req.URL.String())
		}
		body, err := io.ReadAll(req.Body)
		if err != nil {
			t.Fatalf("read body failed: %v", err)
		}
		if !bytes.Contains(body, []byte(`"prompt":"test prompt"`)) {
			t.Fatalf("unexpected request body: %s", string(body))
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body: io.NopCloser(strings.NewReader(
				`{"result_text":"done","current_file":"/tmp/a.mp4"}`,
			)),
			Header: make(http.Header),
		}, nil
	})

	res, err := client.RunCommand(context.Background(), SidecarCommandRequest{Prompt: "test prompt"})
	if err != nil {
		t.Fatalf("RunCommand failed: %v", err)
	}
	if res.ResultText != "done" {
		t.Fatalf("unexpected result text: %q", res.ResultText)
	}
	if res.CurrentFile != "/tmp/a.mp4" {
		t.Fatalf("unexpected current_file: %q", res.CurrentFile)
	}
}

func TestSidecarClientRunCommandReturnsResponseError(t *testing.T) {
	client := NewSidecarClient("http://sidecar.local", 0)
	client.httpClient = newMockHTTPClient(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"error":"bad request"}`)),
			Header:     make(http.Header),
		}, nil
	})

	_, err := client.RunCommand(context.Background(), SidecarCommandRequest{Prompt: "test prompt"})
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "bad request") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSidecarClientRunCommandHTTPError(t *testing.T) {
	client := NewSidecarClient("http://sidecar.local", 0)
	client.httpClient = newMockHTTPClient(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusBadGateway,
			Body:       io.NopCloser(strings.NewReader("down")),
			Header:     make(http.Header),
		}, nil
	})

	_, err := client.RunCommand(context.Background(), SidecarCommandRequest{Prompt: "test prompt"})
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "http 502") {
		t.Fatalf("unexpected error: %v", err)
	}
}
