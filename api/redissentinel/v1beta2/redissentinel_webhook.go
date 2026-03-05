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

package v1beta2

import (
	"context"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const (
	webhookPath = "/validate-redis-redis-opstreelabs-in-v1beta2-redissentinel"
)

// log is for logging in this package.
var redissentinellog = logf.Log.WithName("redissentinel-v1beta2-validation")

// +kubebuilder:webhook:path=/validate-redis-redis-opstreelabs-in-v1beta2-redissentinel,mutating=false,failurePolicy=fail,sideEffects=None,groups=redis.redis.opstreelabs.in,resources=redissentinels,verbs=create;update,versions=v1beta2,name=validate-redissentinel.redis.opstreelabs.in,admissionReviewVersions=v1

// SetupWebhookWithManager will setup the manager
func (r *RedisSentinel) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, &RedisSentinel{}).
		Complete()
}

var _ admission.CustomValidator = &RedisSentinel{}

// ValidateCreate implements admission.CustomValidator so a webhook will be registered for the type
func (r *RedisSentinel) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	sentinel, ok := obj.(*RedisSentinel)
	if !ok {
		return nil, apierrors.NewBadRequest(fmt.Sprintf("expected a RedisSentinel object but got %T", obj))
	}
	redissentinellog.Info("validate create", "name", sentinel.Name)

	return sentinel.validate(nil)
}

// ValidateUpdate implements admission.CustomValidator so a webhook will be registered for the type
func (r *RedisSentinel) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	oldSentinel, ok := oldObj.(*RedisSentinel)
	if !ok {
		return nil, apierrors.NewBadRequest(fmt.Sprintf("expected a RedisSentinel object but got %T", oldObj))
	}
	newSentinel, ok := newObj.(*RedisSentinel)
	if !ok {
		return nil, apierrors.NewBadRequest(fmt.Sprintf("expected a RedisSentinel object but got %T", newObj))
	}
	redissentinellog.Info("validate update", "name", newSentinel.Name)

	return newSentinel.validate(oldSentinel)
}

// ValidateDelete implements admission.CustomValidator so a webhook will be registered for the type
func (r *RedisSentinel) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	sentinel, ok := obj.(*RedisSentinel)
	if !ok {
		return nil, apierrors.NewBadRequest(fmt.Sprintf("expected a RedisSentinel object but got %T", obj))
	}
	redissentinellog.Info("validate delete", "name", sentinel.Name)

	return nil, nil
}

// validate validates the Redis Sentinel CR
func (r *RedisSentinel) validate(_ *RedisSentinel) (admission.Warnings, error) {
	var errors field.ErrorList
	var warnings admission.Warnings

	if r.Spec.Size == nil {
		return warnings, nil
	}

	// Check if the Size is an odd number
	if *r.Spec.Size%2 == 0 {
		errors = append(errors, field.Invalid(
			field.NewPath("spec").Child("clusterSize"),
			*r.Spec.Size,
			"Redis Sentinel cluster size must be an odd number for proper leader election",
		))
	}

	if len(errors) == 0 {
		return nil, nil
	}

	return nil, apierrors.NewInvalid(
		schema.GroupKind{Group: "redis.redis.opstreelabs.in", Kind: "RedisSentinel"},
		r.Name,
		errors,
	)
}

func (r *RedisSentinel) WebhookPath() string {
	return webhookPath
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
