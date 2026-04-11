package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	sqlite "github.com/DeepanshuChaid/Cogito-Ai-Memo.git/internals/database/sqlite"
)

func main() {
    sqlite.InitDB()


    // Target: Anthropic's API (or OpenAI)
    target, _ := url.Parse("https://api.anthropic.com")
    proxy := httputil.NewSingleHostReverseProxy(target)

    // Wrap the director to log requests
    originalDirector := proxy.Director
    proxy.Director = func(req *http.Request) {
        bodyBytes, _ := io.ReadAll(req.Body)
        req.Body = io.NopCloser(bytes.NewReader(bodyBytes))

        originalDirector(req)

        // Log what the AI is trying to do
        log.Printf("[REQUEST] %s %s", req.Method, req.URL.Path)
        log.Printf("[HEADERS] %v", req.Header.Get("X-Request-Id"))
    }

    // Modify response to see what comes back
    proxy.ModifyResponse = func(resp *http.Response) error {
        log.Printf("[RESPONSE] %d %s", resp.StatusCode, resp.Status)
        return nil
    }

    server := &http.Server{
        Addr:    ":8080",
        Handler: proxy,
    }

    log.Println("Cogito proxy listening on :8080")
    log.Println("Forward Claude Code to: http://localhost:8080")
    log.Fatal(server.ListenAndServe())
}
