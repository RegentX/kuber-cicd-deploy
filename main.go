package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	port := env("PORT", "8080")
	imagePath := env("IMAGE_PATH", "/data/image.png")
	deployedBy := env("DEPLOYED_BY", "unknown")

	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("go-image-service\n"))
		w.Write([]byte("deployed-by: " + deployedBy + "\n"))
		w.Write([]byte("try: /image\n"))
	})

	http.HandleFunc("/image", func(w http.ResponseWriter, r *http.Request) {
		data, err := os.ReadFile(imagePath)
		if err != nil {
			http.Error(w, "cannot read image: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "image/png")
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	})

	log.Printf("listening on :%s, image=%s, deployedBy=%s", port, imagePath, deployedBy)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
