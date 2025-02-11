package main

import (
	"fmt"
	"log"
	"net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received request from", r.RemoteAddr)
	fmt.Println(r.Method, r.URL, r.Proto)
	fmt.Println("Host:", r.Host)
	fmt.Println("User-Agent:", r.UserAgent())

	// Respond to the client
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Hello From Backend Server")
	fmt.Println("Replied with a hello message")
}

func main() {
	http.HandleFunc("/", handler)

	port := ":8080" // Backend server runs on port 8081
	fmt.Println("Backend server running on", port)

	log.Fatal(http.ListenAndServe(port, nil))
}
