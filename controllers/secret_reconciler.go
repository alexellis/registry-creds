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

// secretSuffix was: -registrycreds
const secretSuffix = ""

const ignoreAnnotation = "alexellis.io/registry-creds.ignore"

func ignoredNamespace(ns *corev1.Namespace) bool {
	return ns.Annotations[ignoreAnnotation] == "1" || strings.ToLower(ns.Annotations[ignoreAnnotation]) == "true"
}

// Reconcile applies a number of ClusterPullSecrets to ServiceAccounts within
// various valid namespaces. Namespaces can be ignored as required.
func (r *SecretReconciler) Reconcile(clusterPullSecret v1.ClusterPullSecret, ns string) error {
	ctx := context.Background()

	targetNS := &corev1.Namespace{}
	if err := r.Get(ctx, client.ObjectKey{Name: ns}, targetNS); err != nil {
		wrappedErr := errors.Wrapf(err, "unable to fetch namespace: %s", ns)
		r.Log.Info(wrappedErr.Error())
		return wrappedErr
	}

	if ignoredNamespace(targetNS) {
		r.Log.Info(fmt.Sprintf("ignoring namespace %s due to annotation: %s ", ns, ignoreAnnotation))
		return nil
	}

	r.Log.V(10).Info(fmt.Sprintf("Getting SA for: %s", ns))

	if clusterPullSecret.Spec.SecretRef == nil ||
		clusterPullSecret.Spec.SecretRef.Name == "" ||
		clusterPullSecret.Spec.SecretRef.Namespace == "" {
		return fmt.Errorf("no valid secretRef found on ClusterPullSecret: %s.%s",
			clusterPullSecret.Name,
			clusterPullSecret.Namespace)
	}

	pullSecret := &corev1.Secret{}
	if err := r.Get(ctx,
		client.ObjectKey{
			Name:      clusterPullSecret.Spec.SecretRef.Name,
			Namespace: clusterPullSecret.Spec.SecretRef.Namespace},
		pullSecret); err != nil {
		wrappedErr := errors.Wrapf(err, "unable to fetch seedSecret %s.%s", clusterPullSecret.Spec.SecretRef.Name, clusterPullSecret.Spec.SecretRef.Namespace)
		r.Log.Info(fmt.Sprintf("%s", wrappedErr.Error()))
		return wrappedErr
	}

	err := r.createSecret(clusterPullSecret, pullSecret, ns)
	if err != nil {
		r.Log.Info(err.Error())
		return err
	}

	SAs, err := r.listWithin(ns)
	if err != nil {
		wrappedErr := errors.Wrapf(err, "failed to list service accounts in %s namespace", ns)
		r.Log.Info(wrappedErr.Error())
		return wrappedErr
	}

	for _, sa := range SAs.Items {
		err = r.appendSecretToSA(clusterPullSecret, pullSecret, ns, sa.Name)
		if err != nil {
			r.Log.Info(err.Error())
			return err
		}
	}

	return nil
}

func (r *SecretReconciler) listWithin(ns string) (*corev1.ServiceAccountList, error) {
	ctx := context.Background()
	SAs := &corev1.ServiceAccountList{}
	err := r.Client.List(ctx, SAs, client.InNamespace(ns))
	if err != nil {
		return nil, err
	}
	return SAs, nil
}

func (r *SecretReconciler) createSecret(clusterPullSecret v1.ClusterPullSecret, pullSecret *corev1.Secret, ns string) error {
	ctx := context.Background()

	secretKey := clusterPullSecret.Name + secretSuffix

	nsSecret := &corev1.Secret{}
	err := r.Client.Get(ctx, client.ObjectKey{Name: secretKey, Namespace: ns}, nsSecret)
	if err != nil {
		notFound := apierrors.IsNotFound(err)
		if !notFound {
			return errors.Wrap(err, "unexpected error checking for the namespaced pull secret")
		}

		nsSecret = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretKey,
				Namespace: ns,
			},
			Data: pullSecret.Data,
			Type: corev1.SecretTypeDockerConfigJson,
		}

		err = ctrl.SetControllerReference(&clusterPullSecret, nsSecret, r.Scheme)
		if err != nil {
			r.Log.Info(fmt.Sprintf("can't create owner reference: %s.%s, %s", secretKey, ns, err.Error()))
		}

		err = r.Client.Create(ctx, nsSecret)
		if err != nil {
			r.Log.Info(fmt.Sprintf("can't create secret: %s.%s, %s", secretKey, ns, err.Error()))
			return err
		}
		r.Log.Info(fmt.Sprintf("created secret: %s.%s", secretKey, ns))
	}

	return nil
}

func (r *SecretReconciler) appendSecretToSA(clusterPullSecret v1.ClusterPullSecret, pullSecret *corev1.Secret, ns, serviceAccountName string) error {
	ctx := context.Background()

	secretKey := clusterPullSecret.Name + secretSuffix

	sa := &corev1.ServiceAccount{}
	err := r.Client.Get(ctx, client.ObjectKey{Name: serviceAccountName, Namespace: ns}, sa)
	if err != nil {
		if !apierrors.IsConflict(err) {
			r.Log.Info(fmt.Sprintf("error getting SA in namespace: %s, %s", ns, err.Error()))
			wrappedErr := fmt.Errorf("unable to append pull secret to service account: %s", err)
			r.Log.Info(wrappedErr.Error())
			return wrappedErr
		}
		return err
	}

	r.Log.V(10).Info(fmt.Sprintf("Pull secrets: %v", sa.ImagePullSecrets))

	hasSecret := hasImagePullSecret(sa, secretKey)

	if !hasSecret {
		sa.ImagePullSecrets = append(sa.ImagePullSecrets, corev1.LocalObjectReference{
			Name: secretKey,
		})

		err = r.Update(ctx, sa.DeepCopy())
		if err != nil {
			if !apierrors.IsConflict(err) {

				wrappedErr := fmt.Errorf("unable to append pull secret to service account: %s", err)
				r.Log.Info(wrappedErr.Error())
				return err
			}
			return nil
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
