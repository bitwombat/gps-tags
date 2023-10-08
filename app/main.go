package main

import (
	"fmt"
	"net/http"

	"github.com/bitwombat/tag/sub"
)

func main() {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK")
		fmt.Println("Got a health check.")
	})

	http.HandleFunc("/map", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Got a map page request.")
		mapPage, err := sub.GetContents("public_html/index.html", map[string]string{
			"tuckerPositions": `{lat:-31.457435,lng:152.641985}`,
		})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.WriteHeader(http.StatusOK)
		_, err = w.Write([]byte(mapPage))
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
		fmt.Fprintf(w, "OK")

	})
	fs := http.FileServer(http.Dir("./public_html"))
	http.Handle("/", fs)

	fmt.Println("Server running on :80")
	http.ListenAndServe(":80", nil)
}
