package v1beta2

import "fmt"

// GetConnectionInfo returns connection info for clients.
// The dnsDomain parameter should be the cluster DNS domain (e.g., "cluster.local").
func (cr *Redis) GetConnectionInfo(dnsDomain string) *ConnectionInfo {
	return &ConnectionInfo{
		Host: fmt.Sprintf("%s.%s.svc.%s", cr.Name, cr.Namespace, dnsDomain),
		Port: 6379,
	}
}
