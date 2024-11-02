package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

const (
	port          = "4782"
	interfaceName = "pppoe"
	timeout       = 3 * time.Second
	maxAttempts   = 3
)

type Response struct {
	OldIP    string        `json:"old_ip"`
	NewIP    string        `json:"new_ip"`
	Duration time.Duration `json:"duration"`
}

type IPResponse struct {
	Origin string `json:"origin"`
}

var ipServiceURLs = []string{
	"https://wtfismyip.com/text",
	"https://api.ipify.org",
	"https://checkip.amazonaws.com",
	"https://ipecho.net/plain",
	"https://httpbin.org/ip",
	"https://ipnr.dk",
	"https://icanhazip.com",
}

var oldIP string

func getRandomIPServiceURL() string {
	rand.Seed(time.Now().UnixNano())
	return ipServiceURLs[rand.Intn(len(ipServiceURLs))]
}

func getIPAddress() (string, error) {

	client := &http.Client{}

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		url := getRandomIPServiceURL()
		log.Printf("Getting the current IP address... Use %v", url)
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return "", fmt.Errorf("failed to create request: %v", err)
		}

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Attempt %d: failed to obtain IP address: %v", attempt, err)
			if ctx.Err() == context.DeadlineExceeded {
				log.Printf("Attempt %d: request timed out", attempt)
			}
			time.Sleep(500 * time.Millisecond)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("failed to obtain IP address, status code: %d", resp.StatusCode)
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("could not read response body: %v", err)
		}

		var ipResp IPResponse
		if err := json.Unmarshal(body, &ipResp); err == nil {
			cleanedIP := strings.TrimSpace(ipResp.Origin)
			if net.ParseIP(cleanedIP) == nil {
				return "", fmt.Errorf("obtained invalid IP address: %s", cleanedIP)
			}
			log.Printf("Received IP address from JSON: %s", cleanedIP)
			return cleanedIP, nil
		}

		cleanedIP := strings.TrimSpace(string(body))
		if net.ParseIP(cleanedIP) == nil {
			return "", fmt.Errorf("obtained invalid IP address: %s", cleanedIP)
		}
		log.Printf("Received IP address from plain text: %s", cleanedIP)
		return cleanedIP, nil
	}

	return "", fmt.Errorf("failed to obtain IP address after %d attempts", maxAttempts)
}

func isInterfaceUp(interfaceName string) (bool, error) {
	log.Printf("Checking interface status: %s", interfaceName)
	cmd := exec.Command("ifstatus", interfaceName)
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to get interface status: %v", err)
	}

	status := strings.Contains(string(output), `"up": true`)
	log.Printf("Interface status %s: %v", interfaceName, status)
	return status, nil
}

func reconnectInterface(interfaceName string) error {
	log.Printf("Interface reconnection: %s", interfaceName)
	cmd := exec.Command("ifdown", interfaceName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to disable interface: %v", err)
	}

	cmd = exec.Command("ifup", interfaceName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to enable interface: %v", err)
	}

	for {
		isUp, err := isInterfaceUp(interfaceName)
		if err != nil {
			return fmt.Errorf("could not check interface status: %v", err)
		}

		if isUp {
			log.Printf("Interface %s successfully brought up", interfaceName)
			break
		}
		log.Printf("Waiting for interface %s to come up...", interfaceName)
		time.Sleep(1 * time.Second)
	}

	return nil
}

func handler(w http.ResponseWriter, r *http.Request) {
	log.Println("Starting the reconnection process...")
	startTime := time.Now()

	for {
		log.Println("Attempting to reconnect interface...")
		err := reconnectInterface(interfaceName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		time.Sleep(1 * time.Second)

		newIP, err := getIPAddress()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if newIP != oldIP {
			duration := time.Since(startTime)

			response := Response{
				OldIP:    oldIP,
				NewIP:    newIP,
				Duration: duration,
			}
			log.Printf("Success. Old IP: %s", oldIP)
			log.Printf("New IP: %s", newIP)
			log.Printf("IP change duration: %s", duration)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			oldIP = newIP
			return
		}

		log.Printf("IP address has not changed. Old IP: %s, New IP: %s", oldIP, newIP)
	}
}

func main() {
	http.HandleFunc("/", handler)
	log.Printf("Starting server on port %s...", port)

	initialIP, err := getIPAddress()
	if err != nil {
		log.Fatalf("Failed to get initial IP address: %v", err)
	}
	oldIP = initialIP

	log.Fatal(http.ListenAndServe(":"+port, nil))
}
