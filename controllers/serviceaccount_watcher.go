/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	v1 "alexellis/registry-creds/api/v1"
	"context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ServiceAccountWatcher reconciles a ServiceAccount object
type ServiceAccountWatcher struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=core,resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=serviceaccounts/status,verbs=get;update;patch

func (r *ServiceAccountWatcher) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	r.Log.WithValues("serviceaccount", req.NamespacedName)

	var sa corev1.ServiceAccount
	if err := r.Get(ctx, req.NamespacedName, &sa); err != nil {
		r.Log.Info(fmt.Sprintf("%s", errors.Wrap(err, "unable to fetch serviceaccount")))
		return ctrl.Result{}, nil
	}

	r.Log.Info(fmt.Sprintf("detected a change in serviceaccount: %s", sa.Name))

	pullSecretList := &v1.ClusterPullSecretList{}
	err := r.Client.List(ctx, pullSecretList)
	if err != nil {
		r.Log.Info(fmt.Sprintf("unable to list ClusterPullSecrets, %s", err.Error()))
		return ctrl.Result{}, nil
	}

	for _, clusterPullSecret := range pullSecretList.Items {
		err = r.appendSecretToSA(clusterPullSecret, sa.Namespace, sa.Name)
		if err != nil {
			r.Log.Info(err.Error())
			return ctrl.Result{}, nil
		}
	}

	return ctrl.Result{}, nil
}

func (r *ServiceAccountWatcher) appendSecretToSA(clusterPullSecret v1.ClusterPullSecret, ns, serviceAccountName string) error {
	ctx := context.Background()

	secretKey := clusterPullSecret.Name + secretSuffix

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

func (r *ServiceAccountWatcher) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.ServiceAccount{}).
		Complete(r)
}
