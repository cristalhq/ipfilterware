package ipfilterware

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

var dummyHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

func TestHandler(t *testing.T) {
	var count int
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count++
	})

	f, err := New(h, &Config{
		AllowedIPs: []string{"192.0.2.1"},
	})
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("PUT", "http://localhost:7777", http.NoBody)
	f.ServeHTTP(w, r)

	if count != 1 {
		t.Fatalf("want %v, got %v", 1, count)
	}
}

func TestSingleIP(t *testing.T) {
	f, err := New(dummyHandler, &Config{
		AllowedIPs: []string{
			"100.120.130.1/32",
			"200.120.130.1",
			"10.0.0.0/16",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		ip        string
		isAllowed bool
	}{
		{"100.120.130.1", true},
		{"100.120.130.2", false},
		{"10.0.0.1", true},
		{"10.0.30.1", true},
		{"10.20.0.1", false},
	}
	for i, tc := range testCases {
		if f.IsAllowed(net.ParseIP(tc.ip)) != tc.isAllowed {
			t.Errorf("[%d] ip %q must be %v", i, tc.ip, tc.isAllowed)
		}
	}
}

func TestServeHTTP(t *testing.T) {
	const wantCode = http.StatusAccepted
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(wantCode)
	})

	handler, err := New(testHandler, &Config{
		AllowedIPs: []string{"10.20.30.1"},
	})
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("GET", "http://localhost", nil)
	req.RemoteAddr = "10.20.30.1"
	req.RequestURI = ""

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != wantCode {
		t.Fatalf("want %v, got %v", wantCode, resp.StatusCode)
	}
}
