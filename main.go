package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	port := env("PORT", "8080")
	imageURL := env("IMAGE_URL", "https://picsum.photos/800/500")
	deployedBy := env("DEPLOYED_BY", "unknown")

	client := &http.Client{Timeout: 10 * time.Second}

	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("go-image-service\n"))
		w.Write([]byte("deployed-by: " + deployedBy + "\n"))
		w.Write([]byte("image-url: " + imageURL + "\n"))
		w.Write([]byte("try: /image\n"))
	})

	http.HandleFunc("/image", func(w http.ResponseWriter, r *http.Request) {
		req, err := http.NewRequest(http.MethodGet, imageURL, nil)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		resp, err := client.Do(req)
		if err != nil {
			http.Error(w, "failed to fetch image: "+err.Error(), 502)
			return
		}
		defer resp.Body.Close()

		// если удалённый сервер вернул не 200 — пробросим код/текст
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			http.Error(w, "upstream status: "+resp.Status, resp.StatusCode)
			return
		}

		// пробрасываем content-type (если есть)
		if ct := resp.Header.Get("Content-Type"); ct != "" {
			w.Header().Set("Content-Type", ct)
		} else {
			w.Header().Set("Content-Type", "image/*")
		}

		w.WriteHeader(http.StatusOK)
		_, _ = io.Copy(w, resp.Body)
	})

	log.Printf("listening on :%s imageURL=%s deployedBy=%s", port, imageURL, deployedBy)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
