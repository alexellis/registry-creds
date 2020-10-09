package controllers

import (
	"context"
	"fmt"

	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/controller-runtime/pkg/client"

	opsv1 "alexellis/registry-creds/api/v1"
	v1 "alexellis/registry-creds/api/v1"

	corev1 "k8s.io/api/core/v1"
)

// ClusterPullSecretReconciler reconciles a ClusterPullSecret
// object to the default ServiceAccount of each namespace
type ClusterPullSecretReconciler struct {
	client.Client
	Log              logr.Logger
	Scheme           *runtime.Scheme
	SecretReconciler *SecretReconciler
}

// +kubebuilder:rbac:groups=ops.alexellis.io,resources=clusterpullsecrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ops.alexellis.io,resources=clusterpullsecrets/status,verbs=get;update;patch

// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=secrets/status,verbs=get;update;patch

// +kubebuilder:rbac:groups=core,resources=serviceaccounts,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=core,resources=serviceaccounts/status,verbs=get;update;patch

// Reconcile applies a number of ClusterPullSecrets to the default ServiceAccount
// within various valid namespaces. Namespaces can be ignored as required.
func (r *ClusterPullSecretReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	_ = r.Log.WithValues("clusterpullsecret", req.NamespacedName)

	var pullSecret v1.ClusterPullSecret
	if err := r.Get(ctx, req.NamespacedName, &pullSecret); err != nil {
		r.Log.Info(fmt.Sprintf("%s\n", errors.Wrap(err, "unable to fetch pullSecret")))
	} else {

		r.Log.Info(fmt.Sprintf("Found: %s\n", pullSecret.Name))

		serviceaccounts := &corev1.ServiceAccountList{}
		r.Client.List(ctx, serviceaccounts)

		r.Log.Info(fmt.Sprintf("Found %d ServiceAccounts", len(serviceaccounts.Items)))

		for _, serviceaccount := range serviceaccounts.Items {
			namespaced := types.NamespacedName{Name: serviceaccount.Name, Namespace: serviceaccount.Namespace}
			err := r.SecretReconciler.Reconcile(pullSecret, namespaced)
			if err != nil {
				r.Log.Info(fmt.Sprintf("Found error: %s", err.Error()))
			}
		}

	}

	return ctrl.Result{}, nil
}

func (r *ClusterPullSecretReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&opsv1.ClusterPullSecret{}).
		Complete(r)
}
