package ipfilterware

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/netip"
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
	// AllowedIPs list of IPs and/or CIDRs.
	AllowedIPs []string

	// ForbiddenHandler will be invoked when client is blocked.
	// If nil then http.Error will be used.
	ForbiddenHandler http.Handler
}

// New creates a new handler which wraps handler given based on a config.
func New(next http.Handler, cfg *Config) (*Handler, error) {
	h := &Handler{
		next: next,
	}
	if err := h.Update(cfg); err != nil {
		return nil, err
	}
	return h, nil
}

// Wrap a given handler.
func (h *Handler) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		filter := h.ipFilter.Load().(*ipFilter)

		ip, err := ipFromRequest(r)
		if err != nil {
			filter.forbiddenHandler.ServeHTTP(w, r)
			return
		}

		if filter.isAllowed(ip) {
			next.ServeHTTP(w, r)
		} else {
			filter.forbiddenHandler.ServeHTTP(w, r)
		}
	})
}

// ServeHTTP implements http.Handler interface.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	filter := h.ipFilter.Load().(*ipFilter)

	ip, err := ipFromRequest(r)
	if err != nil {
		filter.forbiddenHandler.ServeHTTP(w, r)
		return
	}

	if filter.isAllowed(ip) {
		h.next.ServeHTTP(w, r)
	} else {
		filter.forbiddenHandler.ServeHTTP(w, r)
	}
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

// IsAllowed reports whether given IP is allowed.
func (h *Handler) IsAllowed(ip netip.Addr) bool {
	filter := h.ipFilter.Load().(*ipFilter)
	return filter.isAllowed(ip)
}

// Same as http.Error func.
var defaultForbiddenHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusForbidden)
	fmt.Fprintln(w, http.StatusText(http.StatusForbidden))
})

type ipFilter struct {
	allowedIP        map[netip.Addr]struct{}
	allowedCIDR      []*net.IPNet
	forbiddenHandler http.Handler
}

func newIPFilter(cfg *Config) (*ipFilter, error) {
	if cfg == nil {
		return nil, errors.New("config cannot be nil")
	}

	ipf := &ipFilter{
		forbiddenHandler: defaultForbiddenHandler,
	}
	if cfg.ForbiddenHandler != nil {
		ipf.forbiddenHandler = cfg.ForbiddenHandler
	}

	var err error
	ipf.allowedIP, ipf.allowedCIDR, err = parseIPWithCIDR(cfg.AllowedIPs)
	if err != nil {
		return nil, err
	}
	return ipf, nil
}

func (ipf *ipFilter) isAllowed(ip netip.Addr) bool {
	if !ip.IsValid() {
		return false
	}
	if _, ok := ipf.allowedIP[ip]; ok {
		return true
	}
	for _, cidr := range ipf.allowedCIDR {
		if cidr.Contains(net.ParseIP(ip.String())) {
			return true
		}
	}
	return false
}

func ipFromRequest(r *http.Request) (netip.Addr, error) {
	ip := r.RemoteAddr
	if strings.IndexByte(r.RemoteAddr, byte(':')) >= 0 {
		ip, _, _ = net.SplitHostPort(r.RemoteAddr)
	}
	return netip.ParseAddr(ip)
}

func parseIPWithCIDR(nets []string) (map[netip.Addr]struct{}, []*net.IPNet, error) {
	ips := make(map[netip.Addr]struct{}, len(nets))
	cidrs := make([]*net.IPNet, 0, len(nets))

	for _, n := range nets {
		if _, cidr, err := net.ParseCIDR(n); err == nil {
			cidrs = append(cidrs, cidr)
			continue
		}
		if ip := net.ParseIP(n); ip != nil {
			ips[netip.MustParseAddr(n)] = struct{}{}
			continue
		}
		return nil, nil, fmt.Errorf("bad IP or CIDR: %q", n)
	}
	return ips, cidrs, nil
}
