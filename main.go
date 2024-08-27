package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"time"
)

var (
	centralServerIP string
	nodeId          string
	pingUrls        []string
	mu              sync.Mutex
	loopRunning     bool
)

func init() {
	centralServerIP = os.Getenv("CENTRAL_SERVER_IP")
	if centralServerIP == "" {
		fmt.Println("CENTRAL_SERVER_IP environment variable is required")
		os.Exit(1)
	}
}

func fetchPingUrls() error {
	resp, err := http.Get(fmt.Sprintf("http://%s/register", centralServerIP))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to register and fetch ping URLs: %v", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var response struct {
		NodeId   string   `json:"nodeId"`
		PingUrls []string `json:"pingUrls"`
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return err
	}

	nodeId = response.NodeId
	pingUrls = response.PingUrls

	return nil
}

func ping(url string) (time.Duration, error) {
	start := time.Now()

	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	resp.Body.Close()

	elapsed := fmt.Sprintf("%dms", time.Since(start).Milliseconds())
	return elapsed, nil
}

func sendResults(results map[string]string) error {
	url := fmt.Sprintf("http://%s/collect/%s", centralServerIP, nodeId)
	data, err := json.Marshal(results)
	if err != nil {
		return err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send results: %v", resp.Status)
	}
	return nil
}

func runPingLoop() {
	mu.Lock()
	defer mu.Unlock()

	if loopRunning {
		fmt.Println("Previous loop still running, skipping this round.")
		return
	}

	loopRunning = true
	defer func() { loopRunning = false }()

	results := make(map[string]string)

	for _, url := range pingUrls {
		elapsed, err := ping(url)
		if err != nil {
			fmt.Printf("Failed to ping %s: %v\n", url, err)
			results[url] = "error"
			continue
		}
		fmt.Printf("Ping to %s: %v\n", url, elapsed)
		results[url] = elapsed.String()
	}

	err := sendResults(results)
	if err != nil {
		fmt.Printf("Failed to send results: %v\n", err)
	}
}

func main() {
	err := fetchPingUrls()
	if err != nil {
		fmt.Printf("Failed to fetch ping URLs: %v\n", err)
		return
	}

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			go runPingLoop()
		}
	}
}
