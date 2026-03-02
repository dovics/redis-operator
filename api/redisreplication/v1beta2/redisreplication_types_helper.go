package v1beta2

import (
	"fmt"

	common "github.com/OT-CONTAINER-KIT/redis-operator/api/common/v1beta2"
)

func (cr *RedisReplication) EnableSentinel() bool {
	return cr != nil && cr.Spec.Sentinel != nil && cr.Spec.Sentinel.Size > 0
}

func (cr *RedisReplication) SentinelStatefulSet() string {
	return cr.Name + "-s"
}

func (cr *RedisReplication) RedisStatefulSet() string {
	return cr.Name
}

func (cr *RedisReplication) SentinelHLService() string {
	return cr.Name + "-s-hl"
}

func (cr *RedisReplication) MasterService() string {
	return cr.Name + "-master"
}

// GetConnectionInfo returns connection info for clients based on the mode.
// The dnsDomain parameter should be the cluster DNS domain (e.g., "cluster.local").
func (cr *RedisReplication) GetConnectionInfo(dnsDomain string) *common.ConnectionInfo {
	if cr.EnableSentinel() {
		return &common.ConnectionInfo{
			Host:       fmt.Sprintf("%s.%s.svc.%s", cr.SentinelHLService(), cr.Namespace, dnsDomain),
			Port:       26379,
			MasterName: "mymaster",
		}
	}
	return &common.ConnectionInfo{
		Host: fmt.Sprintf("%s.%s.svc.%s", cr.MasterService(), cr.Namespace, dnsDomain),
		Port: 6379,
	}
}
