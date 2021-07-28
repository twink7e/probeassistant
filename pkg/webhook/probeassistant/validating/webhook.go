package validating

import (
	"context"
	"fmt"
	"github.com/twink7e/probeassistant/api/v1alpha1"
	admissionv1 "k8s.io/api/admission/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// ProbeAssistantValidatingHandler handles ProbeAssistant
type ProbeAssistantValidatingHandler struct {
	// To use the client, you need to do the following:
	// - uncomment it
	// - import sigs.k8s.io/controller-runtime/pkg/client
	// - uncomment the InjectClient method at the bottom of this file.
	Client client.Client

	// Decoder decodes objects
	Decoder *admission.Decoder
}

var _ admission.Handler = &ProbeAssistantValidatingHandler{}

// Handle handles admission requests.
func (h *ProbeAssistantValidatingHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	obj := &v1alpha1.ProbeAssistant{}

	err := h.Decoder.Decode(req, obj)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	var oldSidecarSet *v1alpha1.ProbeAssistant
	//when Operation is update, decode older object
	if req.AdmissionRequest.Operation == admissionv1.Update {
		oldSidecarSet = new(v1alpha1.ProbeAssistant)
		if err := h.Decoder.Decode(
			admission.Request{AdmissionRequest: admissionv1.AdmissionRequest{Object: req.AdmissionRequest.OldObject}},
			oldSidecarSet); err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
	}
	allowed, reason, err := h.validatingProbeAssistant(ctx, obj, oldSidecarSet)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}
	return admission.ValidationResponse(allowed, reason)
}

func (h *ProbeAssistantValidatingHandler) validatingProbeAssistant(ctx context.Context, obj *v1alpha1.ProbeAssistant, order *v1alpha1.ProbeAssistant) (bool, string, error) {
	allErrs := h.validatingProbeAssistantSpec(obj, field.NewPath("spec"))
	if len(allErrs) != 0 {
		return false, "", allErrs.ToAggregate()
	}
	return true, "allowed to be admitted", nil
}

func (h *ProbeAssistantValidatingHandler) validatingProbeAssistantSpec(obj *v1alpha1.ProbeAssistant, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	if obj.Spec.ChangePodPolicy != v1alpha1.ProbeAssistantSpecChangePodPolicyWaitUpdate && obj.Spec.ChangePodPolicy != v1alpha1.ProbeAssistantSpecChangePodPolicyKeepSave {
		allErrs = append(
			allErrs,
			field.Invalid(fldPath,
				obj.Spec.ChangePodPolicy,
				fmt.Sprintf("ChangePodPolicy supports %s %s",
					v1alpha1.ProbeAssistantSpecChangePodPolicyKeepSave,
					v1alpha1.ProbeAssistantSpecChangePodPolicyWaitUpdate)),
		)
	}
	return allErrs
}

var _ inject.Client = &ProbeAssistantValidatingHandler{}

// InjectClient injects the client into the ProbeAssistantValidatingHandler
func (h *ProbeAssistantValidatingHandler) InjectClient(c client.Client) error {
	h.Client = c
	return nil
}

var _ admission.DecoderInjector = &ProbeAssistantValidatingHandler{}

// InjectDecoder injects the decoder into the ProbeAssistantValidatingHandler
func (h *ProbeAssistantValidatingHandler) InjectDecoder(d *admission.Decoder) error {
	h.Decoder = d
	return nil
}
