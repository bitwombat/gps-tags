package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK")
		fmt.Println("Got a health check.")
	})
	fs := http.FileServer(http.Dir("./public_html"))
	http.Handle("/", fs)


	fmt.Println("Server running on :80")
	http.ListenAndServe(":80", nil)
}
