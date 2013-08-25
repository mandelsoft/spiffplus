package dynaml

import (
	"net"
	"strings"

	"github.com/vito/spiff/yaml"
)

type CallExpr struct {
	Name      string
	Arguments []Expression
}

func (e CallExpr) Evaluate(context Context) yaml.Node {
	switch e.Name {
	case "static_ips":
		if len(e.Arguments) == 0 {
			return nil
		}

		nearestNetworkName := context.FindReference([]string{"name"})
		if nearestNetworkName == nil {
			return nil
		}

		networkName, ok := nearestNetworkName.(string)
		if !ok {
			return nil
		}

		network := context.FindFromRoot([]string{"networks", networkName})
		if network == nil {
			return nil
		}

		networkMap, ok := network.(map[string]yaml.Node)
		if !ok {
			return nil
		}

		subnets, ok := networkMap["subnets"]
		if !ok {
			return nil
		}

		subnetsList, ok := subnets.([]yaml.Node)
		if !ok {
			return nil
		}

		// :3
		if len(subnetsList) != 1 {
			return nil
		}

		subnet := subnetsList[0]

		subnetMap, ok := subnet.(map[string]yaml.Node)
		if !ok {
			return nil
		}

		static, ok := subnetMap["static"]
		if !ok {
			return nil
		}

		staticList, ok := static.([]yaml.Node)
		if !ok {
			return nil
		}

		ipPool := []net.IP{}
		for _, ips := range staticList {
			ipsString, ok := ips.(string)
			if !ok {
				continue
			}

			segments := strings.Split(ipsString, "-")
			if len(segments) != 2 {
				continue
			}

			start := net.ParseIP(strings.Trim(segments[0], " "))
			end := net.ParseIP(strings.Trim(segments[1], " "))

			ipPool = append(ipPool, ipRange(start, end)...)
		}

		ips := []yaml.Node{}
		for _, arg := range e.Arguments {
			index := arg.Evaluate(context)
			if index == nil {
				return nil
			}

			i, ok := index.(int)
			if !ok {
				return nil
			}

			if len(ipPool) <= i {
				return nil
			}

			ips = append(ips, ipPool[i].String())
		}

		return ips
	default:
		return nil
	}
}

func ipRange(a, b net.IP) []net.IP {
	prev := a

	ips := []net.IP{a}

	for !prev.Equal(b) {
		next := net.ParseIP(prev.String())
		inc(next)
		ips = append(ips, next)
		prev = next
	}

	return ips
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}
