package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

// tunnel-gateway is a deliberately narrow public edge for local demos. The
// Cloudflare tunnel points here instead of at the full API, so only the
// authenticated viasocket ingestion route is reachable from the internet.
func main() {
	target, err := url.Parse("http://127.0.0.1:8080")
	if err != nil {
		log.Fatal(err)
	}
	proxy := httputil.NewSingleHostReverseProxy(target)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/webhooks/viasocket" {
			http.NotFound(w, r)
			return
		}
		proxy.ServeHTTP(w, r)
	})

	log.Println("[TUNNEL] Restricted webhook gateway listening on http://127.0.0.1:8081")
	log.Fatal(http.ListenAndServe("127.0.0.1:8081", handler))
}
