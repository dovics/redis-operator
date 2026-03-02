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

package redis

import (
	"context"
	"time"

	rvb2 "github.com/OT-CONTAINER-KIT/redis-operator/api/redis/v1beta2"
	"github.com/OT-CONTAINER-KIT/redis-operator/internal/controller/common"
	intctrlutil "github.com/OT-CONTAINER-KIT/redis-operator/internal/controllerutil"
	"github.com/OT-CONTAINER-KIT/redis-operator/internal/envs"
	"github.com/OT-CONTAINER-KIT/redis-operator/internal/k8sutils"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

const (
	RedisFinalizer = "redisFinalizer"
)

// Reconciler reconciles a Redis object
type Reconciler struct {
	client.Client
	K8sClient kubernetes.Interface
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	instance := &rvb2.Redis{}

	err := r.Get(context.TODO(), req.NamespacedName, instance)
	if err != nil {
		return intctrlutil.RequeueECheck(ctx, err, "failed to get redis instance")
	}
	if instance.GetDeletionTimestamp() != nil {
		if err = k8sutils.HandleRedisFinalizer(ctx, r.Client, instance, RedisFinalizer); err != nil {
			return intctrlutil.RequeueE(ctx, err, "failed to handle redis finalizer")
		}
		return intctrlutil.Reconciled()
	}
	if common.ShouldSkipReconcile(ctx, instance) {
		return intctrlutil.Reconciled()
	}
	if err = k8sutils.AddFinalizer(ctx, instance, RedisFinalizer, r.Client); err != nil {
		return intctrlutil.RequeueE(ctx, err, "failed to add finalizer")
	}
	err = k8sutils.CreateStandaloneRedis(ctx, instance, r.K8sClient)
	if err != nil {
		return intctrlutil.RequeueE(ctx, err, "failed to create redis")
	}
	err = k8sutils.CreateStandaloneService(ctx, instance, r.K8sClient)
	if err != nil {
		return intctrlutil.RequeueE(ctx, err, "failed to create service")
	}

	// Update status
	if err = r.reconcileStatus(ctx, instance); err != nil {
		return intctrlutil.RequeueE(ctx, err, "failed to update status")
	}

	return intctrlutil.RequeueAfter(ctx, time.Second*10, "requeue after 10 seconds")
}

// reconcileStatus updates the Redis status with connection information and health state
func (r *Reconciler) reconcileStatus(ctx context.Context, instance *rvb2.Redis) error {
	connectionInfo := instance.GetConnectionInfo(envs.GetServiceDNSDomain())

	// Check StatefulSet health
	stsService := k8sutils.NewStatefulSetService(r.K8sClient)
	isReady := stsService.IsStatefulSetReady(ctx, instance.Namespace, instance.Name)
	readyReplicas := stsService.GetStatefulSetReplicas(ctx, instance.Namespace, instance.Name)

	var state rvb2.RedisState
	if isReady && readyReplicas > 0 {
		state = rvb2.RedisStateReady
	} else {
		state = rvb2.RedisStateCreating
	}

	// Only update if status has changed
	if instance.Status.State == state &&
		instance.Status.ReadyReplicas == readyReplicas &&
		instance.Status.ConnectionInfo != nil &&
		instance.Status.ConnectionInfo.Host == connectionInfo.Host &&
		instance.Status.ConnectionInfo.Port == connectionInfo.Port {
		return nil
	}

	return r.updateStatus(ctx, instance, rvb2.RedisStatus{
		State:          state,
		ReadyReplicas:  readyReplicas,
		ConnectionInfo: connectionInfo,
	})
}

// updateStatus updates the status subresource
func (r *Reconciler) updateStatus(ctx context.Context, redis *rvb2.Redis, status rvb2.RedisStatus) error {
	copy := redis.DeepCopy()
	copy.Spec = rvb2.RedisSpec{}
	copy.Status = status
	return common.UpdateStatus(ctx, r.Client, copy)
}

// SetupWithManager sets up the controller with the Manager.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, opts controller.Options) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&rvb2.Redis{}).
		WithOptions(opts).
		Complete(r)
}
