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
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	appsv1alpha1 "github.com/twink7e/probeassistant/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ProbeAssistantReconciler reconciles a ProbeAssistant object
type ProbeAssistantReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	processor *Processor
}

func NewProbeAssistantReconciler(cli client.Client, sche *runtime.Scheme, rec record.EventRecorder) *ProbeAssistantReconciler {
	return &ProbeAssistantReconciler{
		Client: cli,
		Scheme: sche,
		processor: &Processor{
			Client:   cli,
			recorder: rec,
		},
	}
}

//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=pods/status,verbs=get;update;patch

//+kubebuilder:rbac:groups=apps.k8s.operatoros.io,resources=probeassistants,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps.k8s.operatoros.io,resources=probeassistants/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps.k8s.operatoros.io,resources=probeassistants/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ProbeAssistant object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *ProbeAssistantReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// ctx = context.Background()
	passistannt := &appsv1alpha1.ProbeAssistant{}
	err := r.Get(ctx, req.NamespacedName, passistannt)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}
	klog.V(3).Infof("begin to process ProbeAssistant %v for reconcile", passistannt.Name)

	return r.processor.UpdateProbeAssistant(ctx, passistannt)
}

// SetupWithManager sets up the controller with the Manager.
func (r *ProbeAssistantReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1alpha1.ProbeAssistant{}).
		Complete(r)
}
