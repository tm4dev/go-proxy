package nio

import (
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/phuslu/lru"
	"github.com/rs/zerolog/log"
	"github.com/vlourme/go-proxy/internal/config"
	"github.com/vlourme/go-proxy/internal/utils"
)

var sessions = lru.NewTTLCache[string, net.IP](1024 * 1024)

// GetDialer returns a dialer with a bound IP address.
//
//   - If the session is not found, a new IP address is generated and the session is added to the cache.
//   - If the session is found, the IP address is returned.
//   - If the timeout is not provided, it defaults to 5 minutes.
func GetDialer(ip, session, timeout, location string) (*net.Dialer, error) {
	var local net.IP
	var err error
	var ok bool

	if session == "" {
		local, err = utils.GenerateIP(GetCidrPrefix(location))
		if err != nil {
			return nil, err
		}
	} else {
		local, ok = sessions.Get(session)
		if !ok {
			local, err = utils.GenerateIP(GetCidrPrefix(location))
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

			sessions.Set(session, local, time.Duration(minutes)*time.Minute)
		}
	}

	// Fallback to IPv4 if the target does not match local address family
	if config.Get().EnableFallback && IsIPv6(ip) != IsIPv6(local.String()) {
		fallback := config.GetAnyFallbackPrefix()
		log.Warn().Msgf("IPv4 target, using fallback prefix: %s", fallback)

		return &net.Dialer{
			LocalAddr:     &net.TCPAddr{IP: fallback.IP},
			FallbackDelay: -1,
			Timeout:       5 * time.Second,
			KeepAlive:     -1,
		}, nil
	}

	return &net.Dialer{
		LocalAddr:     &net.TCPAddr{IP: local},
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

func IsIPv6(ip string) bool {
	return strings.Contains(ip, ":")
}
