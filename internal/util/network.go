// internal/util/network.go
package util

import (
	"fmt"
	"net"
	"strings"
)

func GetLocalBaseURL(port int) (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			ip := ipnet.IP.To4()
			if ip != nil && !strings.HasPrefix(ip.String(), "169.254") {
				return fmt.Sprintf("http://%s:%d", ip.String(), port), nil
			}
		}
	}
	return "", fmt.Errorf("could not determine local IP")
}
