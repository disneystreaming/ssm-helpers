package resolver

import (
	"fmt"
	"net"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
)

// Provides a interface for resolving an instance to a single ec2 instance.
type InstanceResolver interface {
	ResolveToInstanceId(client ec2iface.EC2API) ([]string, error)
}

/// HostnameResolver attempts to resolve a fqdn or IP to a corresponding instance ID.
type HostnameResolver struct {
	addrs []string
}

func NewHostnameResolver(addrs []string) *HostnameResolver {
	return &HostnameResolver{
		addrs: addrs,
	}
}

func (hr *HostnameResolver) ResolveToInstanceId(client ec2iface.EC2API) (output []string, err error) {
	ips := make([]*string, 1)
	for _, addr := range hr.addrs {
		ip, err := resolveHostnameToFirst(addr)
		if err != nil {
			return nil, fmt.Errorf("unable to resolve hostname to %v to ip", hr.addrs)
		}

		ipString := ip.String()
		ips = append(ips, &ipString)

	}

	ipFilter := &ec2.Filter{}
	ipFilter.SetName("addresses.private-ip-address").SetValues(ips)

	dniInput := &ec2.DescribeNetworkInterfacesInput{}
	dniInput.SetFilters([]*ec2.Filter{ipFilter})

	describeNetworkInterfacesPager := func(page *ec2.DescribeNetworkInterfacesOutput, lastPage bool) bool {
		for _, nic := range page.NetworkInterfaces {
			output = append(output, *nic.Attachment.InstanceId)
		}

		// If it's not the last page, continue
		return !lastPage
	}

	// Fetch all the instances described
	if err = client.DescribeNetworkInterfacesPages(dniInput, describeNetworkInterfacesPager); err != nil {
		return nil, fmt.Errorf("could not describe network interfaces\n%v", err)
	}

	return output, nil
}

func resolveHostnameToFirst(addr string) (net.IP, error) {
	if ip := net.ParseIP(addr); ip != nil {
		return ip, nil
	} else if ips, err := net.LookupIP(addr); err == nil {
		if len(ips[0]) > 0 {
			return ips[0], nil
		}
	}

	return nil, fmt.Errorf("no IP address found for %s", addr)
}
