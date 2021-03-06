package nebula

import (
	"fmt"
	"net"
	"strconv"
)

type route struct {
	mtu   int
	route *net.IPNet
}

func parseRoutes(config *Config, network *net.IPNet) ([]route, error) {
	var err error

	r := config.Get("tun.routes")
	if r == nil {
		return []route{}, nil
	}

	rawRoutes, ok := r.([]interface{})
	if !ok {
		return nil, fmt.Errorf("tun.routes is not an array")
	}

	if len(rawRoutes) < 1 {
		return []route{}, nil
	}

	routes := make([]route, len(rawRoutes))
	for i, r := range rawRoutes {
		m, ok := r.(map[interface{}]interface{})
		if !ok {
			return nil, fmt.Errorf("entry %v in tun.routes is invalid", i+1)
		}

		rMtu, ok := m["mtu"]
		if !ok {
			return nil, fmt.Errorf("entry %v.mtu in tun.routes is not present", i+1)
		}

		mtu, ok := rMtu.(int)
		if !ok {
			mtu, err = strconv.Atoi(rMtu.(string))
			if err != nil {
				return nil, fmt.Errorf("entry %v.mtu in tun.routes is not an integer: %v", i+1, err)
			}
		}

		if mtu < 500 {
			return nil, fmt.Errorf("entry %v.mtu in tun.routes is below 500: %v", i+1, mtu)
		}

		rRoute, ok := m["route"]
		if !ok {
			return nil, fmt.Errorf("entry %v.route in tun.routes is not present", i+1)
		}

		r := route{
			mtu: mtu,
		}

		_, r.route, err = net.ParseCIDR(fmt.Sprintf("%v", rRoute))
		if err != nil {
			return nil, fmt.Errorf("entry %v.route in tun.routes failed to parse: %v", i+1, err)
		}

		if !ipWithin(network, r.route) {
			return nil, fmt.Errorf(
				"entry %v.route in tun.routes is not contained within the network attached to the certificate; route: %v, network: %v",
				i+1,
				r.route.String(),
				network.String(),
			)
		}

		routes[i] = r
	}

	return routes, nil
}

func ipWithin(o *net.IPNet, i *net.IPNet) bool {
	// Make sure o contains the lowest form of i
	if !o.Contains(i.IP.Mask(i.Mask)) {
		return false
	}

	// Find the max ip in i
	ip4 := i.IP.To4()
	if ip4 == nil {
		return false
	}

	last := make(net.IP, len(ip4))
	copy(last, ip4)
	for x := range ip4 {
		last[x] |= ^i.Mask[x]
	}

	// Make sure o contains the max
	if !o.Contains(last) {
		return false
	}

	return true
}
