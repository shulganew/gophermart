package validators

import (
	"net"
	"net/url"
	"strings"

	"go.uber.org/zap"
)

// Parse server and url address.
func CheckURL(address string) (host string, port string) {
	//Parse address
	link, err := url.Parse(strings.TrimSpace(address))
	if err != nil {
		zap.S().Errorln("Error parsing url: ", err, " return def localhost:8080")
		return "localhost", "8080"
	}

	// Check shema.
	if link.Scheme != "http" {
		//Shema not found, use http
		address = "http://" + address
	}

	link, err = url.Parse(strings.TrimSpace(address))
	if err != nil {
		zap.S().Errorln("Error parsing url whis shema: ", err, " return def localhost:8080")
		return "localhost", "8080"
	}

	// Split address.
	host, port, err = net.SplitHostPort(strings.TrimSpace(link.Host))
	if err != nil {
		zap.S().Errorln("Error split port: ", err, " return def localhost:8080 Host:", link.Host)
		return "", ""
	}
	return host, port
}
