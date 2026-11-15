// Command adapter exposes the customer login lambda over HTTP, for the local ecosystem only.
//
// POST /auth/customer-login is translated into an API Gateway HTTP API (v2) event, invoked on the Lambda Runtime
// Interface Emulator, and the returned APIGatewayV2HTTPResponse is written back (status, headers and body).
// GET /ping answers {"message":"pong"} so the gateway's /ready can check this upstream.
//
// The request body (it carries a CPF) is never logged.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

const maxBodyBytes = 64 << 10

// event is the subset of events.APIGatewayV2HTTPRequest that the lambda reads.
type event struct {
	Version         string            `json:"version"`
	RouteKey        string            `json:"routeKey"`
	RawPath         string            `json:"rawPath"`
	Headers         map[string]string `json:"headers"`
	Body            string            `json:"body"`
	IsBase64Encoded bool              `json:"isBase64Encoded"`
	RequestContext  requestContext    `json:"requestContext"`
}

type requestContext struct {
	RequestID string     `json:"requestId"`
	HTTP      httpDetail `json:"http"`
}

type httpDetail struct {
	Method string `json:"method"`
	Path   string `json:"path"`
}

// response is the subset of events.APIGatewayV2HTTPResponse that the adapter writes back.
type response struct {
	StatusCode      int               `json:"statusCode"`
	Headers         map[string]string `json:"headers"`
	Body            string            `json:"body"`
	IsBase64Encoded bool              `json:"isBase64Encoded"`
}

func main() {
	invokeURL := getenv("LAMBDA_INVOKE_URL", "http://lambda:8080/2015-03-31/functions/function/invocations")
	addr := ":" + getenv("PORT", "9001")

	mux := http.NewServeMux()
	mux.HandleFunc("GET /ping", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"message": "pong"})
	})
	mux.HandleFunc("POST /auth/customer-login", func(w http.ResponseWriter, r *http.Request) {
		login(w, r, invokeURL)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusNotFound, map[string]any{"code": 404, "errors": []string{"route not found"}})
	})

	log.Printf("lambda adapter listening on %s", addr)
	log.Fatal((&http.Server{Addr: addr, Handler: mux, ReadHeaderTimeout: 5 * time.Second}).ListenAndServe())
}

func login(w http.ResponseWriter, r *http.Request, invokeURL string) {
	body, err := io.ReadAll(io.LimitReader(r.Body, maxBodyBytes))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"code": 400, "errors": []string{"invalid body"}})
		return
	}

	requestID := r.Header.Get("X-Request-Id")
	if requestID == "" {
		requestID = fmt.Sprintf("local-%d", time.Now().UnixNano())
	}

	ev := event{
		Version:  "2.0",
		RouteKey: "POST /auth/customer-login",
		RawPath:  r.URL.Path,
		Headers: map[string]string{
			"content-type": r.Header.Get("Content-Type"),
			"x-request-id": requestID,
		},
		Body:           string(body),
		RequestContext: requestContext{RequestID: requestID, HTTP: httpDetail{Method: r.Method, Path: r.URL.Path}},
	}
	payload, err := json.Marshal(ev)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"code": 500, "errors": []string{"internal error"}})
		return
	}

	started := time.Now()
	resp, err := invoke(r, invokeURL, payload)
	if err != nil {
		// The cause is logged without the body; the client only sees a fixed message.
		log.Printf("request_id=%s lambda invoke failed after %dms: %v", requestID, time.Since(started).Milliseconds(), err)
		writeJSON(w, http.StatusBadGateway, map[string]any{"code": 502, "errors": []string{"lambda unavailable"}})
		return
	}
	log.Printf("request_id=%s lambda status=%d duration_ms=%d", requestID, resp.StatusCode, time.Since(started).Milliseconds())

	for k, v := range resp.Headers {
		w.Header().Set(k, v)
	}
	if w.Header().Get("X-Request-Id") == "" {
		w.Header().Set("X-Request-Id", requestID)
	}
	w.WriteHeader(resp.StatusCode)
	_, _ = io.WriteString(w, resp.Body)
}

func invoke(r *http.Request, invokeURL string, payload []byte) (*response, error) {
	req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, invokeURL, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(res.Body, maxBodyBytes))
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("runtime answered %d", res.StatusCode)
	}
	var out response
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("decode lambda response: %w", err)
	}
	if out.IsBase64Encoded {
		return nil, fmt.Errorf("base64 lambda responses are not supported")
	}
	if out.StatusCode == 0 {
		return nil, fmt.Errorf("lambda response without statusCode")
	}
	return &out, nil
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
