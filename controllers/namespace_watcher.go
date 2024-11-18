package controllers

import (
	"context"
	"fmt"

	opsv1 "alexellis/registry-creds/api/v1"

	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"

	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1 "k8s.io/api/core/v1"
)

// NamespaceWatcher watches namespaces for changes to
// trigger ClusterPullSecret reconciliation.
type NamespaceWatcher struct {
	client.Client
	Log              logr.Logger
	Scheme           *runtime.Scheme
	SecretReconciler *SecretReconciler
}

// +kubebuilder:rbac:groups=core,resources=namespaces,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=namespaces/status,verbs=get

func (r *NamespaceWatcher) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.Log.WithValues("namespace", req.NamespacedName)

	var namespace corev1.Namespace
	if err := r.Get(ctx, req.NamespacedName, &namespace); err != nil {
		r.Log.Info(fmt.Sprintf("unable to fetch pullSecret %s, error: %s", req.NamespacedName, err))
		return ctrl.Result{}, nil
	}

	r.Log.V(10).Info(fmt.Sprintf("detected a change in namespace: %s", namespace.Name))

	pullSecretList := &opsv1.ClusterPullSecretList{}
	err := r.Client.List(ctx, pullSecretList)
	if err != nil {
		r.Log.Info(fmt.Sprintf("unable to list ClusterPullSecrets, %s", err.Error()))
		return ctrl.Result{}, nil
	}

	for _, pullSecret := range pullSecretList.Items {
		err := r.SecretReconciler.Reconcile(pullSecret, namespace.Name)
		if err != nil {
			if !errors.IsConflict(err) {
				r.Log.Info(fmt.Sprintf("error reconciling namespace: %s with cluster pull secret: %s, error: %s",
					namespace.Name,
					pullSecret.Name,
					err.Error()))
			}
		}
	}

	return ctrl.Result{}, nil
}

func (r *NamespaceWatcher) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Namespace{}).
		Complete(r)
}
