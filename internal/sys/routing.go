package sys

import (
	"net"

	"github.com/vishvananda/netlink"
)

// AddRoute registers a route to the local interface for the given CIDR
// making it possible to route traffic to the given CIDR.
func AddRoute(cidr string) error {
	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return err
	}

	link, err := netlink.LinkByName("lo")
	if err != nil {
		return err
	}

	return netlink.RouteAdd(&netlink.Route{
		LinkIndex: link.Attrs().Index,
		Dst:       ipnet,
		Protocol:  3,
		Table:     255,
		Scope:     0,
		Priority:  1024,
		Type:      2,
	})
}
