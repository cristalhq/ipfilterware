package ipfilterware

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync/atomic"
)

// Handler to filter client by IP.
type Handler struct {
	next     http.Handler
	ipFilter atomic.Value
}

// Config for the handler.
type Config struct {
	// AllowedIPs list of IPs and/or IP submasks to allow.
	AllowedIPs []string

	// ForbiddenHandler will be invoked when client is blocked.
	// If nil then a default handler will be used.
	ForbiddenHandler http.Handler
}

// New creates a new handler which wraps given based on a config.
func New(next http.Handler, cfg *Config) (*Handler, error) {
	h := &Handler{
		next: next,
	}
	if err := h.Update(cfg); err != nil {
		return nil, err
	}
	return h, nil
}

// ServeHTTP implements http.Handler interface.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ip := ipFromRequest(r)

	filter := h.ipFilter.Load().(*ipFilter)
	if filter.isAllowed(ip) {
		h.next.ServeHTTP(w, r)
		return
	}

	filter.ForbiddenHandler.ServeHTTP(w, r)
}

// Update the handler with a config in a concurrent safe way.
func (h *Handler) Update(cfg *Config) error {
	ipf, err := newIPFilter(cfg)
	if err != nil {
		return err
	}
	h.ipFilter.Store(ipf)
	return nil
}

// Same as http.Error func.
var defaultForbiddenHandler = func(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusForbidden)
	fmt.Fprintln(w, http.StatusText(http.StatusForbidden))
}

type ipFilter struct {
	allowed          []*net.IPNet
	ForbiddenHandler http.Handler
}

func newIPFilter(cfg *Config) (*ipFilter, error) {
	if cfg == nil {
		return nil, errors.New("config cannot be nil")
	}

	ipf := &ipFilter{
		ForbiddenHandler: http.HandlerFunc(defaultForbiddenHandler),
	}
	if cfg.ForbiddenHandler != nil {
		ipf.ForbiddenHandler = cfg.ForbiddenHandler
	}

	var err error
	ipf.allowed, err = parseNets(cfg.AllowedIPs)
	if err != nil {
		return nil, err
	}
	return ipf, nil
}

func (ipf *ipFilter) isAllowed(ip net.IP) bool {
	return isIPInNetwork(ipf.allowed, ip)
}

func isIPInNetwork(nets []*net.IPNet, ip net.IP) bool {
	for _, cidr := range nets {
		if cidr.Contains(ip) {
			return true
		}
	}
	return false
}

func ipFromRequest(r *http.Request) net.IP {
	ip := r.RemoteAddr
	if strings.IndexByte(r.RemoteAddr, byte(':')) >= 0 {
		ip, _, _ = net.SplitHostPort(r.RemoteAddr)
	}
	return net.ParseIP(ip)
}

func parseNets(nets []string) ([]*net.IPNet, error) {
	ipnets := make([]*net.IPNet, 0, len(nets))
	for _, expr := range nets {
		_, cidr, err := net.ParseCIDR(expr)
		if err != nil {
			return nil, err
		}
		ipnets = append(ipnets, cidr)
	}
	return ipnets, nil
}
