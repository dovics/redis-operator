package v1beta2

import (
	common "github.com/OT-CONTAINER-KIT/redis-operator/api/common/v1beta2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type RedisReplicationSpec struct {
	Size                          *int32                            `json:"clusterSize"`
	KubernetesConfig              common.KubernetesConfig           `json:"kubernetesConfig"`
	RedisExporter                 *common.RedisExporter             `json:"redisExporter,omitempty"`
	RedisConfig                   *common.RedisConfig               `json:"redisConfig,omitempty"`
	Storage                       *common.Storage                   `json:"storage,omitempty"`
	NodeSelector                  map[string]string                 `json:"nodeSelector,omitempty"`
	PodSecurityContext            *corev1.PodSecurityContext        `json:"podSecurityContext,omitempty"`
	SecurityContext               *corev1.SecurityContext           `json:"securityContext,omitempty"`
	PriorityClassName             string                            `json:"priorityClassName,omitempty"`
	Affinity                      *corev1.Affinity                  `json:"affinity,omitempty"`
	Tolerations                   *[]corev1.Toleration              `json:"tolerations,omitempty"`
	TLS                           *common.TLSConfig                 `json:"TLS,omitempty"`
	PodDisruptionBudget           *common.RedisPodDisruptionBudget  `json:"pdb,omitempty"`
	ACL                           *common.ACLConfig                 `json:"acl,omitempty"`
	ReadinessProbe                *corev1.Probe                     `json:"readinessProbe,omitempty" protobuf:"bytes,11,opt,name=readinessProbe"`
	LivenessProbe                 *corev1.Probe                     `json:"livenessProbe,omitempty" protobuf:"bytes,12,opt,name=livenessProbe"`
	InitContainer                 *common.InitContainer             `json:"initContainer,omitempty"`
	Sidecars                      *[]common.Sidecar                 `json:"sidecars,omitempty"`
	ServiceAccountName            *string                           `json:"serviceAccountName,omitempty"`
	TerminationGracePeriodSeconds *int64                            `json:"terminationGracePeriodSeconds,omitempty" protobuf:"varint,4,opt,name=terminationGracePeriodSeconds"`
	EnvVars                       *[]corev1.EnvVar                  `json:"env,omitempty"`
	TopologySpreadConstrains      []corev1.TopologySpreadConstraint `json:"topologySpreadConstraints,omitempty"`
	HostPort                      *int                              `json:"hostPort,omitempty"`
	Sentinel                      *Sentinel                         `json:"sentinel,omitempty"`
}

type Sentinel struct {
	common.KubernetesConfig `json:",inline"`
	common.SentinelConfig   `json:",inline"`
	Size                    int32 `json:"size"`
}

func (cr *RedisReplicationSpec) GetReplicationCounts(t string) int32 {
	replica := cr.Size
	return *replica
}

// RedisReplicationState represents the state of a RedisReplication instance
type RedisReplicationState string

const (
	// RedisReplicationStateReady means the Redis replication is ready to serve requests
	RedisReplicationStateReady RedisReplicationState = "Ready"
	// RedisReplicationStateCreating means the Redis replication is being created
	RedisReplicationStateCreating RedisReplicationState = "Creating"
	// RedisReplicationStateConfiguring means the Redis replication is being configured
	RedisReplicationStateConfiguring RedisReplicationState = "Configuring"
	// RedisReplicationStateFailed means the Redis replication has failed
	RedisReplicationStateFailed RedisReplicationState = "Failed"
)

// RedisReplicationStatus defines the observed state of RedisReplication
type RedisReplicationStatus struct {
	// State is the current state of the Redis replication
	// +optional
	State RedisReplicationState `json:"state,omitempty"`
	// ReadyReplicas is the number of ready Redis replicas
	// +optional
	ReadyReplicas int32 `json:"readyReplicas,omitempty"`
	// MasterNode is the name of the current master node
	// +optional
	MasterNode string `json:"masterNode,omitempty"`
	// ConnectionInfo provides connection details for clients to connect to Redis
	// +optional
	ConnectionInfo *common.ConnectionInfo `json:"connectionInfo,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state",description="The current state of the Redis replication",priority=1
// +kubebuilder:printcolumn:name="Desired",type=integer,JSONPath=".spec.clusterSize",description="Desired number of replicas"
// +kubebuilder:printcolumn:name="ReadyReplicas",type=integer,JSONPath=".status.readyReplicas",description="Number of ready replicas"
// +kubebuilder:printcolumn:name="Master",type="string",JSONPath=".status.masterNode",description="Current master node"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description="Age of the Redis replication",priority=1

// Redis is the Schema for the redis API
type RedisReplication struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RedisReplicationSpec   `json:"spec"`
	Status RedisReplicationStatus `json:"status,omitempty"`
}

func (rr *RedisReplication) GetStatefulSetName() string {
	return rr.Name
}

// +kubebuilder:object:root=true

// RedisList contains a list of Redis
type RedisReplicationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RedisReplication `json:"items"`
}

//nolint:gochecknoinits
func init() {
	SchemeBuilder.Register(&RedisReplication{}, &RedisReplicationList{})
}
