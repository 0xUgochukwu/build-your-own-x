package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
)

func handleRequest(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Request received from %s\n", r.RemoteAddr)
	fmt.Printf("%s %s %s\n", r.Method, r.URL, r.Proto)
	for name, headers := range r.Header {
		for _, h := range headers {
			fmt.Printf("%v: %v\n", name, h)
		}
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	r.Body = io.NopCloser(bytes.NewReader(body))

	url := fmt.Sprintf("%s://%s%s", "http", "localhost:8080", r.RequestURI)

	proxyReq, err := http.NewRequest(r.Method, url, bytes.NewReader(body))
	proxyReq.Header = r.Header

	client := &http.Client{}
	res, err := client.Do(proxyReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
	}

	defer res.Body.Close()

	w.WriteHeader(res.StatusCode)
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		http.Error(w, "Failed to read response body", http.StatusInternalServerError)
		return
	}
	_, err = w.Write(resBody)
	if err != nil {
		http.Error(w, "Failed to write response body", http.StatusInternalServerError)
	}

	fmt.Println("Response from server:", res.Proto, res.Status)
	fmt.Println(string(resBody))

}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		go handleRequest(w, r)
	})

	err := http.ListenAndServe(":80", nil)

	if err != nil {
		log.Fatal("Error starting server: ", err)
	}
}
