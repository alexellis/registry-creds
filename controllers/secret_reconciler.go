package controllers

import (
	v1 "alexellis/registry-creds/api/v1"
	"context"
	"fmt"

	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type SecretReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// Reconcile applies an PullSecret to a Namespace
func (r *SecretReconciler) Reconcile(pullSecret v1.ClusterPullSecret, namespaceName string) error {
	ctx := context.Background()

	r.Log.Info(fmt.Sprintf("Getting SA for: %v", namespaceName))

	secretKey := pullSecret.Name + "-registrycreds"

	secret := &corev1.Secret{}
	err := r.Client.Get(ctx, client.ObjectKey{Name: secretKey, Namespace: namespaceName}, secret)
	if err != nil {
		r.Log.Info(fmt.Sprintf("secret not found: %s.%s, %s", secretKey, namespaceName, err.Error()))

		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretKey,
				Namespace: namespaceName,
			},
			Data: pullSecret.Spec.Secret.Data,
			Type: corev1.SecretTypeDockerConfigJson,
		}

		err = r.Client.Create(ctx, secret)
		if err != nil {
			r.Log.Info(fmt.Sprintf("can't create secret: %s.%s, %s", secretKey, namespaceName, err.Error()))
		} else {
			r.Log.Info(fmt.Sprintf("created secret: %s.%s", secretKey, namespaceName))
			err = ctrl.SetControllerReference(&pullSecret, secret, r.Scheme)
			if err != nil {
				r.Log.Info(fmt.Sprintf("can't create owner reference: %s.%s, %s", secretKey, namespaceName, err.Error()))
			}
		}
	}

	sa := &corev1.ServiceAccount{}
	err = r.Client.Get(ctx, client.ObjectKey{Name: "default", Namespace: namespaceName}, sa)
	if err != nil {
		r.Log.Info(fmt.Sprintf("error getting SA in namespace: %s, %s", namespaceName, err.Error()))
	} else {
		r.Log.Info(fmt.Sprintf("Pull secrets: %v", sa.ImagePullSecrets))
		appendSecret := false
		if len(sa.ImagePullSecrets) == 0 {
			appendSecret = true
		} else {
			found := false
			for _, s := range sa.ImagePullSecrets {
				if s.Name == secretKey {
					found = true
				}
			}
			appendSecret = !found
		}

		if appendSecret {
			sa.ImagePullSecrets = append(sa.ImagePullSecrets, corev1.LocalObjectReference{
				Name: secretKey,
			})
			err = r.Update(ctx, sa.DeepCopy())
			if err != nil {
				r.Log.Info(fmt.Sprintf("unable to append pull secret to service account: %s", err))
			}
		}
	}
	return nil
}
