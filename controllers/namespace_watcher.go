package controllers

import (
	"context"
	"fmt"

	opsv1 "alexellis/registry-creds/api/v1"

	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

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
		r.Log.Info(fmt.Sprintf("%s", errors.Wrap(err, "unable to fetch namespace")))
	} else {
		namespaceName := namespace.Name
		r.Log.Info(fmt.Sprintf("detected a change in namespace: %s", namespaceName))

		listOptions := []client.ListOption{
			client.InNamespace(namespaceName),
		}
		saList := &corev1.ServiceAccountList{}
		err = r.List(ctx, saList, listOptions...)
		if err != nil {
			r.Log.Info(fmt.Sprintf("unable to list ServiceAccounts in Namespace: %s, %s", namespaceName, err.Error()))
		} else {

			for _, sa := range saList.Items {
				pullSecretList := &opsv1.ClusterPullSecretList{}
				err := r.Client.List(ctx, pullSecretList)
				if err != nil {
					r.Log.Info(fmt.Sprintf("unable to list ClusterPullSecrets, %s", err.Error()))
				} else {
					for _, pullSecret := range pullSecretList.Items {
						err := r.SecretReconciler.Reconcile(pullSecret, types.NamespacedName{Name: sa.Name, Namespace: sa.Namespace})
						if err != nil {
							r.Log.Info(fmt.Sprintf("error reconciling namespace: %s with cluster pull secret: %s", namespaceName, pullSecret.Name))
						}
					}
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
