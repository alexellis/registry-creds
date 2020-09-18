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

// NamespaceWatcher watches namespaces for changes
type NamespaceWatcher struct {
	client.Client
	Log                 logr.Logger
	Scheme              *runtime.Scheme
	NamespaceReconciler *NamespaceReconciler
}

func (r *NamespaceWatcher) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	_ = r.Log.WithValues("namespace", req.NamespacedName)

	var namespace corev1.Namespace
	if err := r.Get(ctx, req.NamespacedName, &namespace); err != nil {
		r.Log.Info(fmt.Sprintf("%s", errors.Wrap(err, "unable to fetch pullSecret")))
	} else {
		namespaceName := namespace.Name
		r.Log.Info(fmt.Sprintf("detected a change in namespace: %s", namespaceName))

		pullSecretList := &opsv1.ClusterPullSecretList{}
		err := r.Client.List(ctx, pullSecretList)
		if err != nil {
			r.Log.Info(fmt.Sprintf("unable to list ClusterPullSecrets, %s", err.Error()))
		} else {
			for _, pullSecret := range pullSecretList.Items {
				err := r.NamespaceReconciler.Reconcile(pullSecret, namespaceName)
				if err != nil {
					r.Log.Info(fmt.Sprintf("error reconciling namespace: %s with cluster pull secret: %s", namespaceName, pullSecret.Name))
				}
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
