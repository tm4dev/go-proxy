package nio

import (
	"net"
	"strconv"
	"time"

	"github.com/phuslu/lru"
	"github.com/vlourme/go-proxy/internal/config"
	"github.com/vlourme/go-proxy/internal/utils"
)

var sessions = lru.NewTTLCache[string, net.IP](1024 * 1024)

// GetDialer returns a dialer with a bound IP address.
//
//   - If the session is not found, a new IP address is generated and the session is added to the cache.
//   - If the session is found, the IP address is returned.
//   - If the timeout is not provided, it defaults to 5 minutes.
func GetDialer(session, timeout string) (*net.Dialer, error) {
	var ip net.IP
	var err error
	var ok bool

	if session == "" {
		ip, err = utils.GenerateIP(config.GetAnyBindPrefix())
		if err != nil {
			return nil, err
		}
	} else {
		ip, ok = sessions.Get(session)
		if !ok {
			ip, err = utils.GenerateIP(config.GetAnyBindPrefix())
			if err != nil {
				return nil, err
			}

			minutes, err := strconv.Atoi(timeout)
			if err != nil {
				minutes = 5
			}

			if minutes > config.Get().MaxTimeout {
				minutes = config.Get().MaxTimeout
			}

			sessions.Set(session, ip, time.Duration(minutes)*time.Minute)
		}
	}

	return &net.Dialer{
		LocalAddr:     &net.TCPAddr{IP: ip},
		FallbackDelay: -1,
		Timeout:       5 * time.Second,
		KeepAlive:     -1,
	}, nil
}
