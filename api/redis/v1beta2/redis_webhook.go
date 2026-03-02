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
	webhookPath = "/validate-redis-redis-opstreelabs-in-v1beta2-redis"
)

// log is for logging in this package.
var redislog = logf.Log.WithName("redis-v1beta2-validation")

// +kubebuilder:webhook:path=/validate-redis-redis-opstreelabs-in-v1beta2-redis,mutating=false,failurePolicy=fail,sideEffects=None,groups=redis.redis.opstreelabs.in,resources=redis,verbs=create;update,versions=v1beta2,name=validate-redis.redis.opstreelabs.in,admissionReviewVersions=v1

func (r *Redis) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

var _ admission.CustomValidator = &Redis{}

// ValidateCreate implements admission.CustomValidator so a webhook will be registered for the type
func (r *Redis) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	redis, ok := obj.(*Redis)
	if !ok {
		return nil, apierrors.NewBadRequest(fmt.Sprintf("expected a Redis object but got %T", obj))
	}
	redislog.Info("validate create", "name", redis.Name)

	return r.validate(nil)
}

// ValidateUpdate implements admission.CustomValidator so a webhook will be registered for the type
func (r *Redis) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	oldRedis, ok := oldObj.(*Redis)
	if !ok {
		return nil, apierrors.NewBadRequest(fmt.Sprintf("expected a Redis object but got %T", oldObj))
	}
	newRedis, ok := newObj.(*Redis)
	if !ok {
		return nil, apierrors.NewBadRequest(fmt.Sprintf("expected a Redis object but got %T", newObj))
	}
	redislog.Info("validate update", "name", newRedis.Name)

	return r.validate(oldRedis)
}

// ValidateDelete implements admission.CustomValidator so a webhook will be registered for the type
func (r *Redis) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	redis, ok := obj.(*Redis)
	if !ok {
		return nil, apierrors.NewBadRequest(fmt.Sprintf("expected a Redis object but got %T", obj))
	}
	redislog.Info("validate delete", "name", redis.Name)

	return nil, nil
}

// validate validates the Redis CR
func (r *Redis) validate(_ *Redis) (admission.Warnings, error) {
	var errors field.ErrorList

	// Validate ACL configuration
	if r.Spec.ACL != nil {
		if err := r.Spec.ACL.Validate(); err != nil {
			errors = append(errors, field.Invalid(
				field.NewPath("spec").Child("acl"),
				r.Spec.ACL,
				err.Error(),
			))
		}
	}

	if len(errors) == 0 {
		return nil, nil
	}

	return nil, apierrors.NewInvalid(
		schema.GroupKind{Group: "redis.redis.opstreelabs.in", Kind: "Redis"},
		r.Name,
		errors,
	)
}

func (r *Redis) WebhookPath() string {
	return webhookPath
}
