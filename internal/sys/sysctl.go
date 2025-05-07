package sys

import (
	"github.com/lorenzosaino/go-sysctl"
	"github.com/rs/zerolog/log"
)

// TuneSysctl tunes the sysctl settings for the system.
func TuneSysctl() {
	set := func(k, v string) {
		if err := sysctl.Set(k, v); err != nil {
			log.Error().Err(err).Str("key", k).Str("value", v).Msg("sysctl failed")
		}
	}

	// TCP performance
	set("net.ipv4.tcp_fastopen", "3")
	set("net.core.somaxconn", "65535")
	set("net.ipv4.tcp_max_syn_backlog", "65535")
	set("net.core.netdev_max_backlog", "16384")

	// Keepalives
	set("net.ipv4.tcp_keepalive_time", "5")
	set("net.ipv4.tcp_keepalive_intvl", "2")
	set("net.ipv4.tcp_keepalive_probes", "3")

	// Port reuse and time-wait tuning
	set("net.ipv4.tcp_tw_reuse", "1")
	set("net.ipv4.tcp_fin_timeout", "15")
	set("net.ipv4.ip_local_port_range", "1024 65535")

	// Files and connection tracking
	set("fs.file-max", "1048576")

	// IPv6 / proxying
	set("net.ipv6.conf.all.forwarding", "1")
	set("net.ipv6.ip_nonlocal_bind", "1")
	set("net.ipv6.conf.all.disable_ipv6", "0")

	// TCP memory tuning
	set("net.ipv4.tcp_rmem", "4096 87380 6291456")
	set("net.ipv4.tcp_wmem", "4096 65536 6291456")
	set("net.core.rmem_max", "16777216")
	set("net.core.wmem_max", "16777216")
	set("net.ipv4.tcp_congestion_control", "bbr")
}
