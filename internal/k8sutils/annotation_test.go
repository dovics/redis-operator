package k8sutils

import (
	"context"
	"testing"

	rvb2 "github.com/OT-CONTAINER-KIT/redis-operator/api/redis/v1beta2"
	"github.com/OT-CONTAINER-KIT/redis-operator/internal/controller/common"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestCleanupRecreateStatefulsetAnnotation(t *testing.T) {
	scheme := runtime.NewScheme()
	assert.NoError(t, rvb2.AddToScheme(scheme))

	tests := []struct {
		name        string
		initialObj  client.Object
		stsReady    bool
		expectAnnot bool
		expectError bool
	}{
		{
			name: "StatefulSet not ready - annotation should remain",
			initialObj: &rvb2.Redis{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-redis",
					Namespace: "default",
					Annotations: map[string]string{
						common.AnnotationKeyRecreateStatefulset: "true",
					},
				},
			},
			stsReady:    false,
			expectAnnot: true,
			expectError: false,
		},
		{
			name: "StatefulSet ready with annotation - annotation should be removed",
			initialObj: &rvb2.Redis{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-redis",
					Namespace: "default",
					Annotations: map[string]string{
						common.AnnotationKeyRecreateStatefulset:         "true",
						common.AnnotationKeyRecreateStatefulsetStrategy: "Foreground",
					},
				},
			},
			stsReady:    true,
			expectAnnot: false,
			expectError: false,
		},
		{
			name: "StatefulSet ready without annotation - no change",
			initialObj: &rvb2.Redis{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-redis",
					Namespace: "default",
				},
			},
			stsReady:    true,
			expectAnnot: false,
			expectError: false,
		},
		{
			name: "StatefulSet ready with annotation set to false - no change",
			initialObj: &rvb2.Redis{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-redis",
					Namespace: "default",
					Annotations: map[string]string{
						common.AnnotationKeyRecreateStatefulset: "false",
					},
				},
			},
			stsReady:    true,
			expectAnnot: true,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fake client
			fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()

			// Create the object
			ctx := context.Background()
			assert.NoError(t, fakeClient.Create(ctx, tt.initialObj))

			// Call the cleanup function
			err := CleanupRecreateStatefulsetAnnotation(ctx, fakeClient, tt.initialObj, tt.stsReady)

			// Check error expectation
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Get the updated object
			updated := &rvb2.Redis{}
			err = fakeClient.Get(ctx, types.NamespacedName{
				Namespace: tt.initialObj.GetNamespace(),
				Name:      tt.initialObj.GetName(),
			}, updated)
			assert.NoError(t, err)

			// Check annotation expectation
			if tt.expectAnnot {
				assert.NotNil(t, updated.Annotations)
				assert.Equal(t, tt.initialObj.GetAnnotations()[common.AnnotationKeyRecreateStatefulset],
					updated.Annotations[common.AnnotationKeyRecreateStatefulset])
			} else {
				if updated.Annotations != nil {
					assert.Empty(t, updated.Annotations[common.AnnotationKeyRecreateStatefulset])
					assert.Empty(t, updated.Annotations[common.AnnotationKeyRecreateStatefulsetStrategy])
				}
			}
		})
	}
}
