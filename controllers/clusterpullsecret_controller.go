package controllers

import (
	"context"
	"fmt"

	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"

	"sigs.k8s.io/controller-runtime/pkg/client"

	opsv1 "alexellis/registry-creds/api/v1"
	v1 "alexellis/registry-creds/api/v1"

	corev1 "k8s.io/api/core/v1"
)

// ClusterPullSecretReconciler reconciles a ClusterPullSecret object
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

func (r *ClusterPullSecretReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("clusterpullsecret", req.NamespacedName)

	var pullSecret v1.ClusterPullSecret
	if err := r.Get(ctx, req.NamespacedName, &pullSecret); err != nil {
		log.Info(fmt.Sprintf("%s\n", errors.Wrap(err, "unable to fetch pullSecret")))
		return ctrl.Result{}, client.IgnoreNotFound(err)

	}
	log.Info(fmt.Sprintf("Found: %s\n", pullSecret.Name))

	namespaces := &corev1.NamespaceList{}
	r.Client.List(ctx, namespaces)

	log.Info(fmt.Sprintf("Found %d namespaces", len(namespaces.Items)))

	for _, namespace := range namespaces.Items {
		namespaceName := namespace.Name
		err := r.SecretReconciler.Reconcile(pullSecret, namespaceName)
		if err != nil {
			log.Info(fmt.Sprintf("Found error: %s", err.Error()))
		}
	}
	return ctrl.Result{}, nil
}

func (r *ClusterPullSecretReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&opsv1.ClusterPullSecret{}).
		Complete(r)
}
