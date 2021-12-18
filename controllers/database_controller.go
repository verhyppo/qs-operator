/*
Copyright 2021.

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

	corev1 "k8s.io/api/core/v1"
	errorsv1 "k8s.io/apimachinery/pkg/api/errors"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	databasev1alpha1 "github.com/verhyppo/qs-operator/api/v1alpha1"
)

// DatabaseReconciler reconciles a Database object
type DatabaseReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=database.verhyppo.github.io,resources=databases,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=database.verhyppo.github.io,resources=databases/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=database.verhyppo.github.io,resources=databases/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Database object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *DatabaseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)
	ctrl.Log.Info("Reconciling Database")
	// Fetch the Database instance
	instance := &databasev1alpha1.Database{}
	err := r.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	secret := &corev1.Secret{}

	if err := controllerutil.SetControllerReference(instance, secret, r.Scheme); err != nil {
		return reconcile.Result{}, err
	}

	err = r.Client.Get(ctx, types.NamespacedName{Name: instance.Spec.Foo, Namespace: instance.Namespace}, secret)
	if err != nil && errorsv1.IsNotFound(err) {
		// Secret does not exist so we need to fill the one created a few steps ago with relevant data from CR
		err = fillSecretFromCR(instance, secret, r)
		if err != nil { // For some reason the secret could not be created
			return reconcile.Result{}, err
		}
	} else if err != nil { // For some reason the secret could not be fetched
		return reconcile.Result{}, err
	}

	return ctrl.Result{}, nil
}

func fillSecretFromCR(cr *databasev1alpha1.Database, secret *corev1.Secret, r *DatabaseReconciler) error {
	secret.Name = cr.Spec.Foo
	secret.Namespace = cr.Namespace
	secret.Type = corev1.SecretTypeOpaque
	secret.Data = map[string][]byte{
		"username": []byte(cr.Spec.ComplexObject.Username),
		"password": []byte(cr.Spec.ComplexObject.Password),
	}
	return r.Create(context.TODO(), secret)
}

// SetupWithManager sets up the controller with the Manager.
func (r *DatabaseReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&databasev1alpha1.Database{}).
		// Owns(&corev1.Secret{}).
		// Watches(&source.Kind{Type: &databasev1alpha1.Database{}}, &handler.EnqueueRequestForObject{}).
		Complete(r)
}
