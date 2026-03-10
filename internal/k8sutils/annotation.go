package k8sutils

import (
	"context"

	"github.com/OT-CONTAINER-KIT/redis-operator/internal/controller/common"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// CleanupRecreateStatefulsetAnnotation checks if the StatefulSet is ready and
// removes the redis.opstreelabs.in/recreate-statefulset annotation if present.
// This function should be called after the StatefulSet is ready to automatically
// clean up the recreation annotation.
func CleanupRecreateStatefulsetAnnotation(ctx context.Context, cl client.Client, obj client.Object, stsReady bool) error {
	if !stsReady {
		return nil
	}

	annotations := obj.GetAnnotations()
	if annotations == nil {
		return nil
	}

	if annotations[common.AnnotationKeyRecreateStatefulset] != "true" {
		return nil
	}

	// Create a patch to remove the annotation
	patch := client.MergeFrom(obj.DeepCopyObject().(client.Object))
	delete(annotations, common.AnnotationKeyRecreateStatefulset)
	delete(annotations, common.AnnotationKeyRecreateStatefulsetStrategy)
	obj.SetAnnotations(annotations)

	if err := cl.Patch(ctx, obj, patch); err != nil {
		log.FromContext(ctx).Error(err, "failed to remove recreate-statefulset annotation", "object", obj.GetName())
		return err
	}

	log.FromContext(ctx).Info("successfully removed recreate-statefulset annotation after StatefulSet was ready", "object", obj.GetName())
	return nil
}
