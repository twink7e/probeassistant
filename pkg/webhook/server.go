package webhook

import (
	// pamutating "github.com/twink7e/probeassistant/pkg/webhook/probeassistant/mutating"
	// pavalidating "github.com/twink7e/probeassistant/pkg/webhook/probeassistant/validating"
	podmutating "github.com/twink7e/probeassistant/pkg/webhook/pod/mutating"
	// pavalidating "github.com/twink7e/probeassistant/pkg/webhook/probeassistant/validating"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

func SetupWithManager(mgr manager.Manager) error {
	server := mgr.GetWebhookServer()
	server.Host = "0.0.0.0"

	// ProbeAssistant Validate
	// server.Register("/validate-apps-k8s-operatoros-io-v1alpha1-probeassistant", &webhook.Admission{Handler: &pavalidating.ProbeAssistantValidatingHandler{}})
	// ProbeAssistant Mutating
	//server.Register()
	server.Register("/mutate-v1-pod", &webhook.Admission{Handler: &podmutating.PodCreateHandler{}})
	return nil
}
