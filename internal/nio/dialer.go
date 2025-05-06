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
func GetDialer(session, timeout, location string) (*net.Dialer, error) {
	var ip net.IP
	var err error
	var ok bool

	if session == "" {
		ip, err = utils.GenerateIP(GetCidrPrefix(location))
		if err != nil {
			return nil, err
		}
	} else {
		ip, ok = sessions.Get(session)
		if !ok {
			ip, err = utils.GenerateIP(GetCidrPrefix(location))
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

// GetCidrPrefix returns a random CIDR prefix for the given location.
//
//   - If the location is empty or not found, a random bind prefix is returned.
//   - Otherwise, a random prefix from the located prefixes is returned.
func GetCidrPrefix(location string) net.IPNet {
	if location == "" {
		return config.GetAnyBindPrefix()
	}

	prefixes, ok := config.GetLocatedPrefixes()[location]
	if !ok {
		return config.GetAnyBindPrefix()
	}

	return prefixes[utils.RandomInt(len(prefixes))]
}
