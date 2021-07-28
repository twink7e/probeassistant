package controllers

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"

	appsv1alpha1 "github.com/twink7e/probeassistant/api/v1alpha1"
	"github.com/twink7e/probeassistant/pkg/control/probe_assistant"
	"github.com/twink7e/probeassistant/pkg/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type Processor struct {
	Client   client.Client
	recorder record.EventRecorder
}

func (p *Processor) UpdateProbeAssistant(ctx context.Context, pa *appsv1alpha1.ProbeAssistant) (reconcile.Result, error) {
	pods, err := p.getMatchingPods(ctx, pa)
	if err != nil {
		klog.Errorf("ProbeAssistant get matching pods error, err: %v, name: %s", err, pa.Name)
		return reconcile.Result{}, err
	}

	status := calculateStatus(pa, pods)

	if err := p.updateProbeAssistantStatus(ctx, pa, status); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

// If you need update the pod object, you must DeepCopy it
func (p *Processor) getMatchingPods(ctx context.Context, pa *appsv1alpha1.ProbeAssistant) ([]*corev1.Pod, error) {
	// get more faster selector
	selector, err := util.GetFastLabelSelector(pa.Spec.Selector)
	if err != nil {
		return nil, err
	}
	// If ProbeAssistant.Spec.Namespace is empty, then select in cluster
	scopedNamespaces := []string{pa.Spec.Namespace}
	selectedPods, err := p.getSelectedPods(ctx, scopedNamespaces, selector)

	if err != nil {
		return nil, err
	}

	// filter out pods that don't require updated, include the following:
	// 1. Deletion pod
	// 2. Already has binding ProbeAssis
	var filteredPods []*corev1.Pod
	for _, pod := range selectedPods {
		if probe_assistant.IsActivePod(pod) && isPodBindProbeAssistant(pod) {
			filteredPods = append(filteredPods, pod)
		}
	}
	return filteredPods, nil
}

// get selected pods(DisableDeepCopy:true, indicates must be deep copy before update pod objection)
func (p *Processor) getSelectedPods(ctx context.Context, namespaces []string, selector labels.Selector) (relatedPods []*corev1.Pod, err error) {
	// DisableDeepCopy:true, indicates must be deep copy before update pod objection
	listOpts := &client.ListOptions{LabelSelector: selector}
	for _, ns := range namespaces {
		allPods := &corev1.PodList{}
		listOpts.Namespace = ns
		if listErr := p.Client.List(ctx, allPods, listOpts); listErr != nil {
			err = fmt.Errorf("ProbeAssistant list pods by ns error, ns[%s], err:%v", ns, listErr)
			return
		}
		for i := range allPods.Items {
			relatedPods = append(relatedPods, &allPods.Items[i])
		}
	}
	return
}

func (p *Processor) updateProbeAssistantStatus(ctx context.Context, pa *appsv1alpha1.ProbeAssistant, status *appsv1alpha1.ProbeAssistantStatus) error {
	paClone := pa.DeepCopy()
	if err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		paClone.Status = *status

		updateErr := p.Client.Status().Update(ctx, paClone)
		if updateErr == nil {
			return nil
		}

		key := types.NamespacedName{
			Name: paClone.Name,
		}
		if err := p.Client.Get(ctx, key, paClone); err != nil {
			klog.Errorf("error getting updated ProbeAssistant %s from client", paClone.Name)
		}
		return updateErr
	}); err != nil {
		return err
	}
	klog.V(3).Infof("sidecarSet(%s) update status(MatchedPods:%d, ProblemPods:%d) success",
		pa.Name, status.MatchedPods, status.ProblemPods)
	return nil
}

func calculateStatus(pa *appsv1alpha1.ProbeAssistant, pods []*corev1.Pod) *appsv1alpha1.ProbeAssistantStatus {
	// TODO(twink7e) have fix this all status.
	var MatchedPods = int32(len(pods))
	return &appsv1alpha1.ProbeAssistantStatus{
		MatchedPods: MatchedPods,
		ProblemPods: pa.Status.ProblemPods,
	}
}

func isPodBindProbeAssistant(pod *corev1.Pod) bool {
	ns, name := probe_assistant.GetPodBindProbeAssistantName(pod)
	if ns != "" && name != "" {
		return true
	}
	return false
}
