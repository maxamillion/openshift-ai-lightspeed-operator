/*
Copyright 2020 Red Hat
Copyright 2025.

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

package helper

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
)

// Helper is a utility struct that aids in patching Kubernetes objects
type Helper struct {
	client       client.Client
	kclient      kubernetes.Interface
	gvk          schema.GroupVersionKind
	scheme       *runtime.Scheme
	beforeObject client.Object
	before       *unstructured.Unstructured
	after        *unstructured.Unstructured
	changes      map[string]bool
	finalizer    string
	logger       logr.Logger
}

// NewHelper creates a new Helper for the given object
func NewHelper(obj client.Object, crClient client.Client, kclient kubernetes.Interface, scheme *runtime.Scheme, log logr.Logger) (*Helper, error) {
	gvk, err := apiutil.GVKForObject(obj, crClient.Scheme())
	if err != nil {
		return nil, err
	}

	unstructuredObj, err := ToUnstructured(obj)
	if err != nil {
		return nil, err
	}

	return &Helper{
		client:       crClient,
		kclient:      kclient,
		gvk:          gvk,
		scheme:       scheme,
		before:       unstructuredObj,
		beforeObject: obj.DeepCopyObject().(client.Object),
		logger:       log,
		finalizer:    strings.ToLower("openshift-ai.io/" + gvk.Kind),
	}, nil
}

// GetClient returns the controller-runtime client
func (h *Helper) GetClient() client.Client {
	return h.client
}

// GetKClient returns the Kubernetes clientset
func (h *Helper) GetKClient() kubernetes.Interface {
	return h.kclient
}

// GetGKV returns the GroupVersionKind
func (h *Helper) GetGKV() schema.GroupVersionKind {
	return h.gvk
}

// GetScheme returns the runtime scheme
func (h *Helper) GetScheme() *runtime.Scheme {
	return h.scheme
}

// GetAfter returns the unstructured object after changes
func (h *Helper) GetAfter() *unstructured.Unstructured {
	return h.after
}

// GetBefore returns the unstructured object before changes
func (h *Helper) GetBefore() *unstructured.Unstructured {
	return h.before
}

// GetChanges returns a map of changed fields
func (h *Helper) GetChanges() map[string]bool {
	return h.changes
}

// GetBeforeObject returns the original object before changes
func (h *Helper) GetBeforeObject() client.Object {
	return h.beforeObject
}

// GetLogger returns the logger
func (h *Helper) GetLogger() logr.Logger {
	return h.logger
}

// GetFinalizer returns the finalizer string
func (h *Helper) GetFinalizer() string {
	return h.finalizer
}

// SetAfter sets the object state after changes and calculates the diff
func (h *Helper) SetAfter(obj client.Object) error {
	unstructuredObj, err := ToUnstructured(obj)
	if err != nil {
		return err
	}

	h.after = unstructuredObj

	h.changes, err = h.calculateChanges(obj)
	if err != nil {
		return err
	}

	return nil
}

// calculateChanges calculates the differences between before and after
func (h *Helper) calculateChanges(after client.Object) (map[string]bool, error) {
	patch := client.MergeFrom(h.beforeObject)
	diff, err := patch.Data(after)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to calculate patch data")
	}

	patchDiff := map[string]interface{}{}
	if err := json.Unmarshal(diff, &patchDiff); err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal patch data into a map")
	}

	res := make(map[string]bool, len(patchDiff))
	for key := range patchDiff {
		res[key] = true
	}
	return res, nil
}

// PatchInstance patches both metadata and status of an instance
func (h *Helper) PatchInstance(ctx context.Context, instance client.Object) error {
	var err error

	l := log.FromContext(ctx)

	if err = h.SetAfter(instance); err != nil {
		l.Error(err, "Set after and calc patch/diff")
		return err
	}

	changes := h.GetChanges()
	patch := client.MergeFrom(h.GetBeforeObject())

	if changes["metadata"] {
		err = h.GetClient().Patch(ctx, instance, patch)
		if k8s_errors.IsConflict(err) {
			l.Info("Metadata update conflict")
			return err
		} else if err != nil && !k8s_errors.IsNotFound(err) {
			l.Error(err, "Metadata update failed")
			return err
		}
	}

	if changes["status"] {
		err = h.GetClient().Status().Patch(ctx, instance, patch)
		if k8s_errors.IsConflict(err) {
			l.Info("Status update conflict")
			return err

		} else if err != nil && !k8s_errors.IsNotFound(err) {
			l.Error(err, "Status update failed")
			return err
		}
	}
	return nil
}

// ToUnstructured converts a runtime.Object to an unstructured.Unstructured
func ToUnstructured(obj runtime.Object) (*unstructured.Unstructured, error) {
	if _, ok := obj.(runtime.Unstructured); ok {
		obj = obj.DeepCopyObject()
	}
	rawMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return nil, err
	}
	return &unstructured.Unstructured{Object: rawMap}, nil
}
