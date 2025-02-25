package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var (
	servers       []string
	activeServers []string
	serverIndex   uint32
	healthPeriod  time.Duration
	mu            sync.RWMutex
)

func getNextServer() string {
	mu.RLock()
	defer mu.RUnlock()

	if len(activeServers) == 0 {
		return ""
	}

	index := atomic.AddUint32(&serverIndex, 1) - 1
	return activeServers[index%uint32(len(activeServers))]
}

func checkHealth() {
	ticker := time.NewTicker(healthPeriod)
	defer ticker.Stop()

	for {
		<-ticker.C
		mu.Lock()

		healthyServers := []string{}

		var wg sync.WaitGroup

		for _, server := range servers {
			wg.Add(1)

			go func(server string) {
				defer wg.Done()
				res, err := http.Get(server)

				if err == nil && res.StatusCode == http.StatusOK {
					healthyServers = append(healthyServers, server)
				}
			}(server)

		}
		wg.Wait()

		activeServers = healthyServers
		mu.Unlock()
	}
}

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

	server := getNextServer()
	if server == "" {
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
	}

	url := fmt.Sprintf("%s%s", server, r.RequestURI)
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
	hosts := os.Getenv("HOSTS")
	if hosts == "" {
		panic("HOSTS environment variable is required")
	}

	if len(os.Args) != 2 {
		panic("Specify health check period")
	}

	period, e := strconv.Atoi(os.Args[1])
	if e != nil {
		panic("Spicify valid number in seconds for check period")
	}

	servers = strings.Split(hosts, ",")
	healthPeriod = time.Duration(period) * time.Second

	go checkHealth()

	http.HandleFunc("/", handleRequest)

	err := http.ListenAndServe(":80", nil)

	if err != nil {
		log.Fatal("Error starting server: ", err)
	}
}
