package controllers

import (
	v1 "alexellis/registry-creds/api/v1"
	"context"
	"fmt"

	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
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

	if pullSecret.Spec.SecretRef == nil || pullSecret.Spec.SecretRef.Name == "" || pullSecret.Spec.SecretRef.Namespace == "" {
		return fmt.Errorf("no valid secret ref found on ClusterPullSecret: %s.%s", pullSecret.Name, pullSecret.Namespace)
	}

	seedSecret := &corev1.Secret{}
	if err := r.Get(ctx,
		client.ObjectKey{Name: pullSecret.Spec.SecretRef.Name, Namespace: pullSecret.Spec.SecretRef.Namespace},
		seedSecret); err != nil {
		r.Log.Info(fmt.Sprintf("%s", errors.Wrapf(err, "unable to fetch seedSecret %s.%s", pullSecret.Spec.SecretRef.Name, pullSecret.Spec.SecretRef.Namespace)))
	} else {

		nsSecret := &corev1.Secret{}
		err := r.Client.Get(ctx, client.ObjectKey{Name: secretKey, Namespace: namespaceName}, nsSecret)
		if err != nil {
			if apierrors.IsNotFound(err) {

				r.Log.Info(fmt.Sprintf("secret not found: %s.%s, %s", secretKey, namespaceName, err.Error()))

				nsSecret = &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      secretKey,
						Namespace: namespaceName,
					},
					Data: seedSecret.Data,
					Type: corev1.SecretTypeDockerConfigJson,
				}
				err = ctrl.SetControllerReference(&pullSecret, nsSecret, r.Scheme)
				if err != nil {
					r.Log.Info(fmt.Sprintf("can't create owner reference: %s.%s, %s", secretKey, namespaceName, err.Error()))
				}

				err = r.Client.Create(ctx, nsSecret)
				if err != nil {
					r.Log.Info(fmt.Sprintf("can't create secret: %s.%s, %s", secretKey, namespaceName, err.Error()))
				} else {
					r.Log.Info(fmt.Sprintf("created secret: %s.%s", secretKey, namespaceName))

				}
			}
		} else {
			return errors.Wrap(err, "unexpected error checking for the namespaced pull secret")
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
	}
	return nil
}
