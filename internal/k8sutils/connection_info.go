/*
Copyright 2020 Opstree Solutions.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package k8sutils

import (
	"context"

	commonapi "github.com/OT-CONTAINER-KIT/redis-operator/api/common/v1beta2"
	rvb2 "github.com/OT-CONTAINER-KIT/redis-operator/api/redis/v1beta2"
	rcvb2 "github.com/OT-CONTAINER-KIT/redis-operator/api/rediscluster/v1beta2"
	rrvb2 "github.com/OT-CONTAINER-KIT/redis-operator/api/redisreplication/v1beta2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// GetRedisConnectionInfo retrieves connection information from the actual service
func GetRedisConnectionInfo(ctx context.Context, client kubernetes.Interface, redis *rvb2.Redis, dnsDomain string) *commonapi.ConnectionInfo {
	// Try to get the additional service first (for LoadBalancer/NodePort types)
	serviceName := redis.Name + "-additional"
	service, err := client.CoreV1().Services(redis.Namespace).Get(ctx, serviceName, metav1.GetOptions{})

	// If additional service doesn't exist, try the main service
	if err != nil {
		if errors.IsNotFound(err) {
			serviceName = redis.Name
			service, err = client.CoreV1().Services(redis.Namespace).Get(ctx, serviceName, metav1.GetOptions{})
		}
		if err != nil {
			log.FromContext(ctx).V(1).Info("Service not found, using basic connection info", "service", serviceName, "error", err.Error())
			// Fall back to basic connection info
			return redis.GetConnectionInfo(dnsDomain)
		}
	}

	// Build ConnectionInfo from the actual service
	connInfo := &commonapi.ConnectionInfo{
		Host:      getServiceHost(service, redis.Namespace, dnsDomain),
		Port:      getServicePort(service, redis.Namespace),
		Type:      string(service.Spec.Type),
		ClusterIP: service.Spec.ClusterIP,
		Domain:    dnsDomain,
	}

	// Add external IPs if present
	if len(service.Spec.ExternalIPs) > 0 {
		connInfo.ExternalIPs = service.Spec.ExternalIPs
	}

	// Add ports from the service
	if len(service.Spec.Ports) > 0 {
		connInfo.Ports = make([]commonapi.ServicePort, 0, len(service.Spec.Ports))
		for _, port := range service.Spec.Ports {
			connInfo.Ports = append(connInfo.Ports, commonapi.ServicePort{
				Name:     port.Name,
				Port:     port.Port,
				Protocol: string(port.Protocol),
			})
		}
	}

	// Add LoadBalancer ingress IPs if present
	if service.Spec.Type == corev1.ServiceTypeLoadBalancer && len(service.Status.LoadBalancer.Ingress) > 0 {
		for _, ingress := range service.Status.LoadBalancer.Ingress {
			if ingress.IP != "" {
				connInfo.ExternalIPs = append(connInfo.ExternalIPs, ingress.IP)
			}
			if ingress.Hostname != "" {
				connInfo.ExternalIPs = append(connInfo.ExternalIPs, ingress.Hostname)
			}
		}
	}

	return connInfo
}

// GetRedisReplicationConnectionInfo retrieves connection information from the actual service
func GetRedisReplicationConnectionInfo(ctx context.Context, client kubernetes.Interface, rr *rrvb2.RedisReplication, dnsDomain string) *commonapi.ConnectionInfo {
	var serviceName string

	// Determine service name based on mode
	if rr.EnableSentinel() {
		serviceName = rr.SentinelHLService()
	} else {
		serviceName = rr.MasterService()
	}

	service, err := client.CoreV1().Services(rr.Namespace).Get(ctx, serviceName, metav1.GetOptions{})
	if err != nil {
		log.FromContext(ctx).V(1).Info("Service not found, using basic connection info", "service", serviceName, "error", err.Error())
		// Fall back to basic connection info
		return rr.GetConnectionInfo(dnsDomain)
	}

	// Build ConnectionInfo from the actual service
	connInfo := &commonapi.ConnectionInfo{
		Host:      getServiceHost(service, rr.Namespace, dnsDomain),
		Port:      getServicePort(service, rr.Namespace),
		Type:      string(service.Spec.Type),
		ClusterIP: service.Spec.ClusterIP,
		Domain:    dnsDomain,
	}

	// Add MasterName for Sentinel mode
	if rr.EnableSentinel() {
		connInfo.MasterName = "mymaster"
	}

	// Add external IPs if present
	if len(service.Spec.ExternalIPs) > 0 {
		connInfo.ExternalIPs = service.Spec.ExternalIPs
	}

	// Add ports from the service
	if len(service.Spec.Ports) > 0 {
		connInfo.Ports = make([]commonapi.ServicePort, 0, len(service.Spec.Ports))
		for _, port := range service.Spec.Ports {
			connInfo.Ports = append(connInfo.Ports, commonapi.ServicePort{
				Name:     port.Name,
				Port:     port.Port,
				Protocol: string(port.Protocol),
			})
		}
	}

	// Add LoadBalancer ingress IPs if present
	if service.Spec.Type == corev1.ServiceTypeLoadBalancer && len(service.Status.LoadBalancer.Ingress) > 0 {
		for _, ingress := range service.Status.LoadBalancer.Ingress {
			if ingress.IP != "" {
				connInfo.ExternalIPs = append(connInfo.ExternalIPs, ingress.IP)
			}
			if ingress.Hostname != "" {
				connInfo.ExternalIPs = append(connInfo.ExternalIPs, ingress.Hostname)
			}
		}
	}

	return connInfo
}

// getServiceHost returns the appropriate host for connecting to the service
func getServiceHost(service *corev1.Service, namespace, dnsDomain string) string {
	// For LoadBalancer or NodePort with external IPs, use those
	if service.Spec.Type == corev1.ServiceTypeLoadBalancer && len(service.Status.LoadBalancer.Ingress) > 0 {
		ingress := service.Status.LoadBalancer.Ingress[0]
		if ingress.IP != "" {
			return ingress.IP
		}
		if ingress.Hostname != "" {
			return ingress.Hostname
		}
	}

	// For ExternalIPs
	if len(service.Spec.ExternalIPs) > 0 {
		return service.Spec.ExternalIPs[0]
	}

	// Default to service DNS name
	return getServiceDNSName(service.Name, namespace, dnsDomain)
}

// getServicePort returns the appropriate Redis port from the service
func getServicePort(service *corev1.Service, namespace string) int {
	// Look for the redis-client or sentinel-client port
	for _, port := range service.Spec.Ports {
		if port.Name == "redis-client" || port.Name == "sentinel-client" {
			return int(port.Port)
		}
	}

	// Fallback to first port or default
	if len(service.Spec.Ports) > 0 {
		return int(service.Spec.Ports[0].Port)
	}

	return 6379 // Default Redis port
}

// getServiceDNSName returns the DNS name of the service
func getServiceDNSName(serviceName, namespace, dnsDomain string) string {
	return serviceName + "." + namespace + ".svc." + dnsDomain
}

// GetRedisClusterConnectionInfo retrieves connection information from the actual service
func GetRedisClusterConnectionInfo(ctx context.Context, client kubernetes.Interface, rc *rcvb2.RedisCluster, dnsDomain string) *commonapi.ConnectionInfo {
	// RedisCluster uses leader service as the primary connection point
	serviceName := rc.Name + "-leader"
	service, err := client.CoreV1().Services(rc.Namespace).Get(ctx, serviceName, metav1.GetOptions{})
	if err != nil {
		log.FromContext(ctx).V(1).Info("Service not found, using basic connection info", "service", serviceName, "error", err.Error())
		// Fall back to basic connection info
		return &commonapi.ConnectionInfo{
			Host: getServiceDNSName(serviceName, rc.Namespace, dnsDomain),
			Port: 6379,
		}
	}

	// Build ConnectionInfo from the actual service
	connInfo := &commonapi.ConnectionInfo{
		Host:      getServiceHost(service, rc.Namespace, dnsDomain),
		Port:      getServicePort(service, rc.Namespace),
		Type:      string(service.Spec.Type),
		ClusterIP: service.Spec.ClusterIP,
		Domain:    dnsDomain,
	}

	// Add external IPs if present
	if len(service.Spec.ExternalIPs) > 0 {
		connInfo.ExternalIPs = service.Spec.ExternalIPs
	}

	// Add ports from the service
	if len(service.Spec.Ports) > 0 {
		connInfo.Ports = make([]commonapi.ServicePort, 0, len(service.Spec.Ports))
		for _, port := range service.Spec.Ports {
			connInfo.Ports = append(connInfo.Ports, commonapi.ServicePort{
				Name:     port.Name,
				Port:     port.Port,
				Protocol: string(port.Protocol),
			})
		}
	}

	// Add LoadBalancer ingress IPs if present
	if service.Spec.Type == corev1.ServiceTypeLoadBalancer && len(service.Status.LoadBalancer.Ingress) > 0 {
		for _, ingress := range service.Status.LoadBalancer.Ingress {
			if ingress.IP != "" {
				connInfo.ExternalIPs = append(connInfo.ExternalIPs, ingress.IP)
			}
			if ingress.Hostname != "" {
				connInfo.ExternalIPs = append(connInfo.ExternalIPs, ingress.Hostname)
			}
		}
	}

	return connInfo
}
