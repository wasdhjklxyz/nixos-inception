package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const configPath = "/etc/nixos-inception/config"

func main() {
	addr, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	url := fmt.Sprintf("http://%s", strings.TrimSpace(string(addr)))

	client := &http.Client{
		Timeout: 30 * time.Second, /* TODO: Make configurable */
	}

	for {
		resp, err := client.Get(url)
		if err != nil {
			time.Sleep(5 * time.Second) /* TODO: Make configurable */
			continue
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		fmt.Println(string(body))
	}
}
