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

package v1alpha1

import (
	"fmt"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var probeassistantlog = logf.Log.WithName("probeassistant-resource")

func (r *ProbeAssistant) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-apps-k8s-operatoros-io-v1alpha1-probeassistant,mutating=true,failurePolicy=fail,sideEffects=None,groups=apps.k8s.operatoros.io,resources=probeassistants,verbs=create;update,versions=v1alpha1,name=mprobeassistant.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Defaulter = &ProbeAssistant{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *ProbeAssistant) Default() {
	probeassistantlog.Info("default", "name", r.Name)

	if r.Spec.ChangePodPolicy == "" {
		r.Spec.ChangePodPolicy = ProbeAssistantSpecChangePodPolicyWaitUpdate
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-apps-k8s-operatoros-io-v1alpha1-probeassistant,mutating=false,failurePolicy=fail,sideEffects=None,groups=apps.k8s.operatoros.io,resources=probeassistants,verbs=create;update;delete,versions=v1alpha1,name=vprobeassistant.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Validator = &ProbeAssistant{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *ProbeAssistant) ValidateCreate() error {
	probeassistantlog.Info("validate create", "name", r.Name)

	return r.ValidateProbeAssistant()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *ProbeAssistant) ValidateUpdate(old runtime.Object) error {
	probeassistantlog.Info("validate update", "name", r.Name)

	return r.ValidateProbeAssistant()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *ProbeAssistant) ValidateDelete() error {
	probeassistantlog.Info("validate delete", "name", r.Name)

	return nil
}

func (r *ProbeAssistant) ValidateProbeAssistant() error {
	var allErrs field.ErrorList
	if err := r.ValidateSpecUpdatePolicy(field.NewPath("spec")); err != nil {
		allErrs = append(allErrs, err)
	}
	if len(allErrs) == 0 {
		return nil
	}
	return apierrors.NewInvalid(
		schema.GroupKind{Group: "apps.k8s.operatoros.io", Kind: "ProbeAssistant"},
		r.Name, allErrs)
}

func (r *ProbeAssistant) ValidateSpecUpdatePolicy(fldPath *field.Path) *field.Error {
	if r.Spec.ChangePodPolicy != ProbeAssistantSpecChangePodPolicyWaitUpdate && r.Spec.ChangePodPolicy != ProbeAssistantSpecChangePodPolicyKeepSave {
		return field.Invalid(fldPath,
			r.Spec.ChangePodPolicy,
			fmt.Sprintf("ChangePodPolicy supports %s %s",
				ProbeAssistantSpecChangePodPolicyKeepSave,
				ProbeAssistantSpecChangePodPolicyWaitUpdate))
	}
	return nil
}
