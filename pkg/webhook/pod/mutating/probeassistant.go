package mutating

import (
	"context"
	appsv1alpha1 "github.com/twink7e/probeassistant/api/v1alpha1"
	pactrl "github.com/twink7e/probeassistant/pkg/control/probeassistant"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// mutate pod based on SidecarSet Object
func (h *PodCreateHandler) probeAssistantMutatingPod(ctx context.Context, req admission.Request, pod *corev1.Pod) error {
	if len(req.AdmissionRequest.SubResource) > 0 ||
		req.AdmissionRequest.Operation != admissionv1.Create ||
		req.AdmissionRequest.Resource.Resource != "pods" {
		return nil
	}
	// filter out pods that don't require inject, include the following:
	// 1. Deletion pod
	if !pactrl.IsActivePod(pod) {
		return nil
	}

	// DisableDeepCopy:true, indicates must be deep copy before update sidecarSet objection
	probeassistantList := &appsv1alpha1.ProbeAssistantList{}
	if err := h.Client.List(ctx, probeassistantList, &client.ListOptions{}); err != nil {
		return err
	}
	var matchedProbeAssistan *appsv1alpha1.ProbeAssistant
	for _, pa := range probeassistantList.Items {
		if matched, err := pactrl.PodMatchedProbeAssistant(pod, &pa); err != nil {
			return err
		} else if matched {
			matchedProbeAssistan = &pa
			break
		}
	}
	if matchedProbeAssistan == nil {
		return nil
	}

	klog.V(3).Infof("[ProbeAssistant inject] begin to operation(%s) pod(%s/%s) resources(%s) subResources(%s)",
		req.Operation, req.Namespace, req.Name, req.Resource, req.SubResource)

	return h.buildProbeAssistant(ctx, pod, matchedProbeAssistan)

}

func (h *PodCreateHandler) buildProbeAssistant(ctx context.Context, pod *corev1.Pod, pa *appsv1alpha1.ProbeAssistant) error {
	// filter containers
	containersName := pactrl.GetBindingContainers(pod)
	if len(*containersName) < 0 {
		klog.V(3).Infof(
			"ProbeAssistant: Pod(%s.%s) did not match any container name, please check pod's annotation(%s).",
			pod.Namespace,
			pod.Name,
			pactrl.ProbeAssistantBindingContainersAnnotation)
		return nil
	}
	// get matched container's idx from Pod.Spec.Containers
	matchedContainerIndex := pactrl.GetMatchedContainerIndex(pod, containersName)

	// set bingding(set Annotations)
	if err := pactrl.MakeProbeAssistantBingdingPod(matchedContainerIndex, pod, pa); err != nil {
		return nil
	}

	// inject/transfer pod liveness and readiness.
	if err := pactrl.InjectPodProbe(pod, matchedContainerIndex); err != nil {
		return err
	}

	// inject container mount
	if err := pactrl.InjectMount(ctx, h.Client, pa, pod, matchedContainerIndex); err != nil {
		return err
	}

	return nil
}
