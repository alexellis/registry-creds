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

// ServiceAccountWatcher watches serviceaccounts for changes
type ServiceAccountWatcher struct {
	client.Client
	Log              logr.Logger
	Scheme           *runtime.Scheme
	SecretReconciler *SecretReconciler
}

// +kubebuilder:rbac:groups=core,resources=serviceaccount,verbs=get;list;watch;create;update;patch;delete

func (r *ServiceAccountWatcher) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	_ = r.Log.WithValues("serviceaccount", req.NamespacedName)

	var serviceaccount corev1.ServiceAccount
	if err := r.Get(ctx, req.NamespacedName, &serviceaccount); err != nil {
		r.Log.Info(fmt.Sprintf("%s", errors.Wrap(err, "unable to fetch serviceaccount")))
	} else {
		r.Log.Info(fmt.Sprintf("detected a change in serviceaccount: %s", req.NamespacedName))

		pullSecretList := &opsv1.ClusterPullSecretList{}
		err := r.Client.List(ctx, pullSecretList)
		if err != nil {
			r.Log.Info(fmt.Sprintf("unable to list ClusterPullSecrets, %s", err.Error()))
		} else {
			for _, pullSecret := range pullSecretList.Items {
				err := r.SecretReconciler.Reconcile(pullSecret, req.NamespacedName)
				if err != nil {
					r.Log.Info(fmt.Sprintf("error reconciling serviceaccount: %s with cluster pull secret: %s", req.NamespacedName, pullSecret.Name))
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
