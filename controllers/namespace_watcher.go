package controllers

import (
	"context"
	"fmt"

	opsv1 "alexellis/registry-creds/api/v1"

	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
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

// +kubebuilder:rbac:groups=core,resources=namespaces,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=namespaces/status,verbs=get;update;patch

func (r *NamespaceWatcher) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	r.Log.WithValues("namespace", req.NamespacedName)

	var namespace corev1.Namespace
	if err := r.Get(ctx, req.NamespacedName, &namespace); err != nil {
		r.Log.Info(fmt.Sprintf("%s", errors.Wrap(err, "unable to fetch pullSecret")))
		return ctrl.Result{}, nil
	}

	r.Log.Info(fmt.Sprintf("detected a change in namespace: %s", namespace.Name))

	pullSecretList := &opsv1.ClusterPullSecretList{}
	err := r.Client.List(ctx, pullSecretList)
	if err != nil {
		r.Log.Info(fmt.Sprintf("unable to list ClusterPullSecrets, %s", err.Error()))
		return ctrl.Result{}, nil
	}

	for _, pullSecret := range pullSecretList.Items {
		err := r.SecretReconciler.Reconcile(pullSecret, namespace.Name)
		if err != nil {
			r.Log.Info(fmt.Sprintf("error reconciling namespace: %s with cluster pull secret: %s, error: %s",
				namespace.Name,
				pullSecret.Name,
				err.Error()))
		}
	}

	return ctrl.Result{}, nil
}

func (r *NamespaceWatcher) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Namespace{}).
		Complete(r)
}
