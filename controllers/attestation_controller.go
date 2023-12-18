/*
Copyright 2023.

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
	"context"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	keylimev1alpha1 "github.com/sarroutbi/osdk-attestation-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

// AttestationReconciler reconciles a Attestation object
type AttestationReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=keylime.redhat.com,resources=attestations,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=keylime.redhat.com,resources=attestations/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=keylime.redhat.com,resources=attestations/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;update
//+kubebuilder:rbac:groups=core,resources=pods/status,verbs=get;list;watch;create;update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Attestation object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *AttestationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	SetLogInstance(log.FromContext(ctx))
	a := &keylimev1alpha1.Attestation{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: req.NamespacedName.Namespace,
			Name:      req.NamespacedName.Name,
		},
	}
	err := r.Get(ctx, req.NamespacedName, a)
	if err != nil {
		if errors.IsNotFound(err) {
			GetLogInstance().Info("Attestation resource not found")
		}
	}
	r.CheckSpec(a, ctx)
	r.VersionUpdate(a)
	err = r.Client.Status().Update(context.Background(), a)
	if err != nil {
		GetLogInstance().Error(err, "Unable to update Attestation status")
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *AttestationReconciler) VersionUpdate(attestation *keylimev1alpha1.Attestation) {
	v := new(VersionUpdater)
	v.NewVersionUpdater(attestation)
	v.UpdateVersion()
}

func (r *AttestationReconciler) CheckSpec(attestation *keylimev1alpha1.Attestation, ctx context.Context) error {
	GetLogInstance().Info("Checking Pod List", "Spec", attestation.Spec)
	if attestation.Spec.PodRetrievalInfo != nil && attestation.Spec.PodRetrievalInfo.Enabled {
		// TODO: Set namespace in CRD
		lpods, e := PodList(attestation.Spec.PodRetrievalInfo.Namespace, ctx)
		GetLogInstance().Info("Logging Pod List", "Pod List", lpods, "Error", e)
		attestation.Status.PodList = lpods
	} else {
		GetLogInstance().Info("Pod List not retrieved")
		attestation.Status.PodList = nil
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AttestationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&keylimev1alpha1.Attestation{}).
		Complete(r)
}
