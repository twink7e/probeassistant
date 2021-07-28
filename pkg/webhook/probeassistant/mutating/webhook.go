package mutating

import "sigs.k8s.io/controller-runtime/pkg/webhook/admission"

// SidecarSetCreateHandler handles SidecarSet
type ProbeAssistantCreateHandler struct {
	// Decoder decodes objects
	Decoder *admission.Decoder
}
