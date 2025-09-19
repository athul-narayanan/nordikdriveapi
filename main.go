package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"google.golang.org/genai"
)

func handler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	_, err := genai.NewClient(ctx, nil)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create Gemini client: %v", err), http.StatusInternalServerError)
		return
	}

	fmt.Fprintln(w, "Gemini client created successfully!")
}

func main() {
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
