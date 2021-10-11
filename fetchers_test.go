package ipfilterware_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
)

// Copy-paste friendly fetchers.

// Doer represents HTTP client.
type Doer interface {
	Do(req *http.Request) (resp *http.Response, err error)
}

// FetchCloudflareIPv4 from https://www.cloudflare.com/ips-v4
func FetchCloudflareIPv4(ctx context.Context, client Doer) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://www.cloudflare.com/ips-v4", nil)
	if err != nil {
		return nil, err
	}
	return fetchAndSplit(client, req)
}

// FetchCloudflareIPv6 from https://www.cloudflare.com/ips-v6
func FetchCloudflareIPv6(ctx context.Context, client Doer) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://www.cloudflare.com/ips-v6", nil)
	if err != nil {
		return nil, err
	}
	return fetchAndSplit(client, req)
}

func fetchAndSplit(client Doer, req *http.Request) ([]string, error) {
	body, err := fetch(client, req)
	if err != nil {
		return nil, err
	}

	lines := bytes.Split(body, []byte{'\n'})
	ips := make([]string, 0, len(lines))
	for _, line := range lines {
		ips = append(ips, string(line))
	}
	return ips, nil
}

func fetch(client Doer, req *http.Request) ([]byte, error) {
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}
