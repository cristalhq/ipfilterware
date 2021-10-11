package ipfilterware_test

import (
	"context"
	"net/http"
	"time"

	"github.com/cristalhq/ipfilterware"
)

func ExampleFetchCloudflare() {
	ctx := context.Background()
	ips, err := FetchCloudflareIPv4(ctx, &http.Client{Timeout: 5 * time.Second})
	if err != nil {
		panic(err)
	}

	var myHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// do something good
	})

	handler, err := ipfilterware.New(myHandler, &ipfilterware.Config{
		AllowedIPs: ips,
	})
	if err != nil {
		panic(err)
	}

	// Use handler as a middleware in your router
	_ = handler
}
