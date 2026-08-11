package middleware

import (
	"net"
	"net/http"
)

func WithTrustedSubnet(cidr string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cidr == "" {
				next.ServeHTTP(w, r)
				return
			}

			_, subnet, err := net.ParseCIDR(cidr)
			if err != nil {
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}

			ipStr := r.Header.Get("X-Real-IP")
			if ipStr == "" {
				http.Error(w, "X-Real-IP header is missing", http.StatusForbidden)
				return
			}

			ip := net.ParseIP(ipStr)
			if ip == nil {
				http.Error(w, "invalid X-Real-IP header", http.StatusForbidden)
				return
			}

			if !subnet.Contains(ip) {
				http.Error(w, "IP not in trusted subnet", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
