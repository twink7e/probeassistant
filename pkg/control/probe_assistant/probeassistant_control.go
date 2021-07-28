package probe_assistant

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	appsv1alpha1 "github.com/twink7e/probeassistant/api/v1alpha1"
)

// PodMatchSidecarSet determines if pod match Selector of ProbeAssistant.
func PodMatchedProbeAssistant(pod *corev1.Pod, pa *appsv1alpha1.ProbeAssistant) (bool, error) {
	//If matchedNamespace is not empty, ProbeAssistant will only match the pods in the namespace
	if pa.Spec.Namespace != "" && pa.Spec.Namespace != pod.Namespace {
		return false, nil
	}
	// if selector not matched, then continue
	selector, err := metav1.LabelSelectorAsSelector(pa.Spec.Selector)
	if err != nil {
		return false, err
	}

	if !selector.Empty() && selector.Matches(labels.Set(pod.Labels)) {
		return true, nil
	}
	return false, nil
}
