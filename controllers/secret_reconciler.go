package controllers

import (
	v1 "alexellis/registry-creds/api/v1"
	"context"
	"fmt"
	"strings"

	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// SecretReconciler adds a secret to the default
// ServiceAccount in each namespace, unless the namespace
// has an ignore annotation
type SecretReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// Reconcile applies an PullSecret to a Namespace
func (r *SecretReconciler) Reconcile(pullSecret v1.ClusterPullSecret, namespacedName types.NamespacedName) error {
	ctx := context.Background()
	const ignoreAnnotation = "alexellis.io/registry-creds.ignore"
	const includePullSecretAnnotation = "alexellis.io/registry-creds.include"

	targetNamespace := &corev1.Namespace{}
	if err := r.Get(ctx, client.ObjectKey{Name: namespacedName.Namespace}, targetNamespace); err != nil {
		wrappedErr := errors.Wrapf(err, "unable to fetch namespace for introspection:%s", namespacedName.Namespace)
		r.Log.Info(wrappedErr.Error())
		return wrappedErr
	}

	if targetNamespace.Annotations[ignoreAnnotation] == "1" || strings.ToLower(targetNamespace.Annotations[ignoreAnnotation]) == "true" {
		r.Log.Info(fmt.Sprintf("ignoring serviceaccount %s due to namespace annotation %v present ", namespacedName, ignoreAnnotation))
		return nil
	}

	r.Log.Info(fmt.Sprintf("Getting ServiceAccount: %v", namespacedName.Namespace))
	secretKey := pullSecret.Name + "-registrycreds"

	if clusterPullSecret.Spec.SecretRef == nil ||
		clusterPullSecret.Spec.SecretRef.Name == "" ||
		clusterPullSecret.Spec.SecretRef.Namespace == "" {
		return fmt.Errorf("no valid secretRef found on ClusterPullSecret: %s.%s",
			clusterPullSecret.Name,
			clusterPullSecret.Namespace)
	}

	pullSecret := &corev1.Secret{}
	if err := r.Get(ctx,
		client.ObjectKey{Name: pullSecret.Spec.SecretRef.Name, Namespace: pullSecret.Spec.SecretRef.Namespace},
		seedSecret); err != nil {
		if apierrors.IsNotFound(err) {
			r.Log.Info(fmt.Sprintf("%s", errors.Wrapf(err, "seedSecret not found %s/%s", pullSecret.Spec.SecretRef.Name, pullSecret.Spec.SecretRef.Namespace)))
		} else {
			r.Log.Info(fmt.Sprintf("%s", errors.Wrapf(err, "unable to fetch seedSecret: %s/%s", pullSecret.Spec.SecretRef.Name, pullSecret.Spec.SecretRef.Namespace)))
		}
	} else {

		nsSecret := &corev1.Secret{}
		err := r.Client.Get(ctx, client.ObjectKey{Name: secretKey, Namespace: targetNamespace.Name}, nsSecret)
		if err != nil {
			if apierrors.IsNotFound(err) {

				r.Log.Info(fmt.Sprintf("secret not found: %s.%s, %s", secretKey, targetNamespace.Name, err.Error()))

				nsSecret = &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      secretKey,
						Namespace: targetNamespace.Name,
					},
					Data: seedSecret.Data,
					Type: corev1.SecretTypeDockerConfigJson,
				}
				err = ctrl.SetControllerReference(&pullSecret, nsSecret, r.Scheme)
				if err != nil {
					r.Log.Info(fmt.Sprintf("can't create owner reference: %s.%s, %s", secretKey, targetNamespace.Name, err.Error()))
				}

				err = r.Client.Create(ctx, nsSecret)
				if err != nil {
					r.Log.Info(fmt.Sprintf("can't create secret: %s.%s, %s", secretKey, targetNamespace.Name, err.Error()))
				} else {
					r.Log.Info(fmt.Sprintf("created secret: %s.%s", secretKey, targetNamespace.Name))

				}
			} else {
				return errors.Wrap(err, "unexpected error checking for the namespaced pull secret")
			}
		}

		sa := &corev1.ServiceAccount{}
		err = r.Client.Get(ctx, client.ObjectKey{Name: namespacedName.Name, Namespace: namespacedName.Namespace}, sa)
		r.Log.Info(fmt.Sprintf("%s's Pull secrets: %v", sa.Name, sa.ImagePullSecrets))

		appendSecret := false
		if sa.Annotations[includePullSecretAnnotation] == pullSecret.Name || sa.Name == "default" {
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
		} else {
			r.Log.Info(fmt.Sprintf("ignoring serviceaccount: %s, not default or annotation (%v) missing.", sa.Name, includePullSecretAnnotation))
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

func (r *SecretReconciler) appendSecretToSA(clusterPullSecret v1.ClusterPullSecret, pullSecret *corev1.Secret, ns, serviceAccountName string) error {
	ctx := context.Background()

	secretKey := clusterPullSecret.Name + "-registrycreds"

	sa := &corev1.ServiceAccount{}
	err := r.Client.Get(ctx, client.ObjectKey{Name: serviceAccountName, Namespace: ns}, sa)
	if err != nil {
		r.Log.Info(fmt.Sprintf("error getting SA in namespace: %s, %s", ns, err.Error()))
		wrappedErr := fmt.Errorf("unable to append pull secret to service account: %s", err)
		r.Log.Info(wrappedErr.Error())
		return wrappedErr
	}

	r.Log.Info(fmt.Sprintf("Pull secrets: %v", sa.ImagePullSecrets))

	hasSecret := hasImagePullSecret(sa, secretKey)

	if !hasSecret {
		sa.ImagePullSecrets = append(sa.ImagePullSecrets, corev1.LocalObjectReference{
			Name: secretKey,
		})

		err = r.Update(ctx, sa.DeepCopy())
		if err != nil {
			wrappedErr := fmt.Errorf("unable to append pull secret to service account: %s", err)
			r.Log.Info(wrappedErr.Error())
			return err
		}
	}

	return nil
}

func hasImagePullSecret(sa *corev1.ServiceAccount, secretKey string) bool {
	found := false
	if len(sa.ImagePullSecrets) > 0 {
		for _, s := range sa.ImagePullSecrets {
			if s.Name == secretKey {
				found = true
				break
			}
		}
	}
	return found
}
