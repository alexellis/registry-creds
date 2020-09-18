package controllers

import (
	v1 "alexellis/registry-creds/api/v1"
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ServiceAccountReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// Reconcile applies a PullSecret to a ServiceAccount
func (r *ServiceAccountReconciler) Reconcile(pullSecret v1.ClusterPullSecret, serviceAccountName string, namespaceName string) error {
	ctx := context.Background()
	targetServiceAccount := &corev1.ServiceAccount{}
	if err := r.Get(ctx, client.ObjectKey{Name: serviceAccountName, Namespace: namespaceName}, targetServiceAccount); err != nil {
		wrappedErr := errors.Wrapf(err, "ServiceAccountReconciler unable to fetch service account for introspection:%s", serviceAccountName)
		r.Log.Info(wrappedErr.Error())
		return wrappedErr
	}

	r.Log.Info(fmt.Sprintf("ServiceAccountReconciler Getting SA for: %v", serviceAccountName))
	secretKey := pullSecret.Name + "-credentials"

	if pullSecret.Spec.SecretRef == nil || pullSecret.Spec.SecretRef.Name == "" || pullSecret.Spec.SecretRef.Namespace == "" {
		return fmt.Errorf("ServiceAccountReconciler no valid secret ref found on ClusterPullSecret: %s.%s", pullSecret.Name, pullSecret.Namespace)
	}

	r.Log.Info(fmt.Sprintf("ServiceAccountReconciler Pull secrets: %v", targetServiceAccount.ImagePullSecrets))
	appendSecret := false
	if len(targetServiceAccount.ImagePullSecrets) == 0 {
		appendSecret = true
	} else {
		found := false
		for _, s := range targetServiceAccount.ImagePullSecrets {
			if s.Name == secretKey {
				found = true
			}
		}
		appendSecret = !found
	}

	if appendSecret {
		targetServiceAccount.ImagePullSecrets = append(targetServiceAccount.ImagePullSecrets, corev1.LocalObjectReference{
			Name: secretKey,
		})
		err := r.Update(ctx, targetServiceAccount.DeepCopy())
		if err != nil {
			r.Log.Info(fmt.Sprintf("ServiceAccountReconciler unable to append pull secret to service account: %s", err))
		}
	}

	return nil
}
