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

// ServiceAccountWatcher watches service accounts for changes
type ServiceAccountWatcher struct {
	client.Client
	Log                      logr.Logger
	Scheme                   *runtime.Scheme
	ServiceAccountReconciler *ServiceAccountReconciler
}

func (r *ServiceAccountWatcher) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	_ = r.Log.WithValues("ServiceAccountWatcher service account", req.NamespacedName)

	var serviceaccount corev1.ServiceAccount
	if err := r.Get(ctx, req.NamespacedName, &serviceaccount); err != nil {
		r.Log.Info(fmt.Sprintf("%s", errors.Wrap(err, "ServiceAccountWatcher unable to fetch pullSecret")))
	} else {
		serviceaccountName := serviceaccount.Name
		r.Log.Info(fmt.Sprintf("ServiceAccountWatcher detected a change in service account: %s", serviceaccountName))

		pullSecretList := &opsv1.ClusterPullSecretList{}
		err := r.Client.List(ctx, pullSecretList)
		if err != nil {
			r.Log.Info(fmt.Sprintf("ServiceAccountWatcher unable to list ClusterPullSecrets, %s", err.Error()))
		} else {
			for _, pullSecret := range pullSecretList.Items {
				err := r.ServiceAccountReconciler.Reconcile(pullSecret, serviceaccountName, req.Namespace)
				if err != nil {
					r.Log.Info(fmt.Sprintf("ServiceAccountWatcher error reconciling service account: %s with cluster pull secret: %s", serviceaccountName, pullSecret.Name))
				}
			}
		}
	}

	return ctrl.Result{}, nil
}

func (r *ServiceAccountWatcher) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.ServiceAccount{}).
		Complete(r)
}
