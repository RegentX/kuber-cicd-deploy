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
	imagePath := env("IMAGE_PATH", "/data/image.png")
	deployedBy := env("DEPLOYED_BY", "unknown")

	client := &http.Client{Timeout: 10 * time.Second}

	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("go-image-service\n"))
		_, _ = w.Write([]byte("deployed-by: " + deployedBy + "\n"))
		_, _ = w.Write([]byte("image-url: " + imageURL + "\n"))
		_, _ = w.Write([]byte("image-path: " + imagePath + "\n"))
		_, _ = w.Write([]byte("try: /image (local) or /proxy (remote)\n"))
	})

	// ЛОКАЛЬНОЕ — 100% будет работать без интернета
	http.HandleFunc("/image", func(w http.ResponseWriter, r *http.Request) {
		f, err := os.Open(imagePath)
		if err != nil {
			http.Error(w, "failed to open local image: "+err.Error(), 500)
			return
		}
		defer f.Close()

		// у тебя png в образе — можно явно
		w.Header().Set("Content-Type", "image/png")
		w.WriteHeader(http.StatusOK)
		_, _ = io.Copy(w, f)
	})

	// УДАЛЁННОЕ — если кластеру разрешён egress
	http.HandleFunc("/proxy", func(w http.ResponseWriter, r *http.Request) {
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

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			http.Error(w, "upstream status: "+resp.Status, resp.StatusCode)
			return
		}

		if ct := resp.Header.Get("Content-Type"); ct != "" {
			w.Header().Set("Content-Type", ct)
		} else {
			w.Header().Set("Content-Type", "image/*")
		}

		w.WriteHeader(http.StatusOK)
		_, _ = io.Copy(w, resp.Body)
	})

	log.Printf("listening on :%s imageURL=%s imagePath=%s deployedBy=%s", port, imageURL, imagePath, deployedBy)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
