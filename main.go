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

	body, err := io.ReadAll(resp.Body)  // Updated to use io.ReadAll
	if err != nil {
		return err
	}

	// Expecting a response in the form of:
	// { "nodeId": "some-id", "pingUrls": ["url1", "url2", ...] }
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
	cmd := exec.Command("curl", "-o", "/dev/null", "-s", "-w", "%{time_total}", url)
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	elapsed, err := time.ParseDuration(fmt.Sprintf("%sms", string(output)))
	if err != nil {
		return 0, err
	}
	return elapsed, nil
}

func sendResults(data string) error {
	url := fmt.Sprintf("http://%s/collect/%s", centralServerIP, nodeId)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer([]byte(data)))
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

	results := "{"
	for _, url := range pingUrls {
		elapsed, err := ping(url)
		if err != nil {
			fmt.Printf("Failed to ping %s: %v\n", url, err)
			results += fmt.Sprintf("\"%s\":\"error\",", url)
			continue
		}
		fmt.Printf("Ping to %s: %v\n", url, elapsed)
		results += fmt.Sprintf("\"%s\":\"%v\",", url, elapsed)
	}
	results = results[:len(results)-1] + "}" // Remove last comma and close JSON

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