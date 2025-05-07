package config

import (
	"flag"
	"net"
	"os"

	"github.com/goccy/go-yaml"
	"github.com/rs/zerolog/log"
	"github.com/vlourme/go-proxy/internal/utils"
)

type AuthType string

const (
	AuthTypeNone        AuthType = "none"
	AuthTypeCredentials AuthType = "credentials"
	AuthTypeRedis       AuthType = "redis"
)

// Config is the configuration for the proxy.
type Config struct {
	// ListenAddress is the address to listen on.
	ListenAddress string `yaml:"listen_address"`
	// ListenPort is the port to listen on.
	ListenPort uint16 `yaml:"listen_port"`
	// DebugMode is whether to enable debug mode.
	DebugMode bool `yaml:"debug_mode"`
	// TestPort is the port to test the proxy.
	TestPort uint16 `yaml:"test_port"`
	// MaxTimeout is the maximum timeout for a session.
	MaxTimeout int `yaml:"max_timeout"`
	// HTTPClose is whether to force "Connection: close" header in HTTP-only requests.
	HTTPClose bool `yaml:"http_close"`
	// Auth is the authentication configuration.
	Auth struct {
		Type        AuthType `yaml:"type"`
		Credentials struct {
			Username string `yaml:"username"`
			Password string `yaml:"password"`
		} `yaml:"credentials"`
		Redis struct {
			DSN string `yaml:"dsn"`
		} `yaml:"redis"`
	} `yaml:"auth"`
	// BindPrefixes is the list of prefixes to bind to.
	BindPrefixes []string `yaml:"bind_prefixes"`
	// EnableFallback is whether to enable the fallback prefix.
	EnableFallback bool `yaml:"enable_fallback"`
	// FallbackPrefixes is the list of IPv4 prefixes to bind if the target does not support IPv6.
	FallbackPrefixes []string `yaml:"fallback_prefixes"`
	// LocatedPrefixes is the list of prefixes to bind to for each location.
	LocatedPrefixes map[string][]string `yaml:"located_prefixes"`
	// ReplaceIPs is the list of IPs to replace with the override.
	ReplaceIPs map[string]string `yaml:"replace_ips"`
	// DeletedHeaders is the list of headers to delete.
	DeletedHeaders []string `yaml:"deleted_headers"`
}

var config *Config
var bindPrefixes = []net.IPNet{}
var fallbackPrefixes = []net.IPNet{}
var locatedPrefixes = map[string][]net.IPNet{}
var replaceIPs = map[*net.IPNet]string{}

func load() *Config {
	path := flag.String("config", "config.yaml", "The path to the config file")
	flag.Parse()

	yamlFile, err := os.ReadFile(*path)
	if err != nil {
		log.Fatal().Err(err).Msg("Error reading config file")
	}

	var cfg Config
	err = yaml.Unmarshal(yamlFile, &cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Error parsing config file")
	}

	for _, prefix := range cfg.BindPrefixes {
		_, ipnet, err := net.ParseCIDR(prefix)
		if err != nil {
			log.Fatal().Err(err).Msg("Error parsing bind prefix")
		}
		bindPrefixes = append(bindPrefixes, *ipnet)
	}

	for _, prefix := range cfg.FallbackPrefixes {
		_, ipnet, err := net.ParseCIDR(prefix)
		if err != nil {
			log.Fatal().Err(err).Msg("Error parsing fallback prefix")
		}
		fallbackPrefixes = append(fallbackPrefixes, *ipnet)
	}

	for location, prefixes := range cfg.LocatedPrefixes {
		for _, prefix := range prefixes {
			_, ipnet, err := net.ParseCIDR(prefix)
			if err != nil {
				log.Fatal().Err(err).Msg("Error parsing located prefix")
			}
			locatedPrefixes[location] = append(locatedPrefixes[location], *ipnet)
		}
	}

	for cidr, ip := range cfg.ReplaceIPs {
		_, ipnet, err := net.ParseCIDR(cidr)
		if err != nil {
			log.Fatal().Err(err).Msg("Error parsing replace IP")
		}
		replaceIPs[ipnet] = ip
	}

	return &cfg
}

// Get returns the parsed config
func Get() *Config {
	if config == nil {
		config = load()
	}

	return config
}

// GetBindPrefixes returns the bind prefixes
func GetBindPrefixes() []net.IPNet {
	return bindPrefixes
}

// GetAnyBindPrefix returns a random bind prefix
func GetAnyBindPrefix() net.IPNet {
	return bindPrefixes[utils.RandomInt(len(bindPrefixes))]
}

// GetFallbackPrefixes returns the fallback prefixes
func GetFallbackPrefixes() []net.IPNet {
	return fallbackPrefixes
}

// GetAnyFallbackPrefix returns a random fallback prefix
func GetAnyFallbackPrefix() net.IPNet {
	return fallbackPrefixes[utils.RandomInt(len(fallbackPrefixes))]
}

// GetLocatedPrefixes returns the located prefixes
func GetLocatedPrefixes() map[string][]net.IPNet {
	return locatedPrefixes
}

// GetReplaceIPs returns the replace IPs
func GetReplaceIPs() map[*net.IPNet]string {
	return replaceIPs
}
