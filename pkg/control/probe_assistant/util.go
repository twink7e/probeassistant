package probe_assistant

import (
	"context"
	"encoding/json"
	appsv1alpha1 "github.com/twink7e/probeassistant/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

const (
	ProbeAssistantNameAnnotation                = "apps.k8s.operatoros.io/probeassiant-name"
	ProbeAssistantBindingContainersAnnotation   = "apps.k8s.operatoros.io/probeassiant-binding-containers"
	ProbeAssistantNameChangePodPolicyAnnotation = "apps.k8s.operatoros.io/probeassiant-change-pod-policy"
	ProbeAssistantContainerStatusMapAnnotation  = "apps.k8s.operatoros.io/probeassiant-container-status"
)

type ProbeAssistantContainerStatusMap map[string]ProbeAssistantContainerStatus

type ProbeAssistantContainerStatus struct {
	Name              string      `json:"name"`
	StatusOfLiveness  string      `json:"statusOfLiveness"`
	StatusOfReadiness string      `json:"statusOfReadiness"`
	UpdateTimestamp   metav1.Time `json:"updateTimestamp"`
}

// IsActivePod determines the pod whether need be injected and updated
func IsActivePod(pod *corev1.Pod) bool {
	if pod.ObjectMeta.GetDeletionTimestamp() != nil {
		return false
	}
	return true
}

func GetContainerProbeStatus(pod metav1.Object, containerName string) *ProbeAssistantContainerStatus {
	hashKey := ProbeAssistantContainerStatusMapAnnotation
	annotations := pod.GetAnnotations()
	if annotations[hashKey] == "" {
		return nil
	}
	containerProbeStatus := &ProbeAssistantContainerStatus{}
	if err := json.Unmarshal([]byte(annotations[hashKey]), containerProbeStatus); err != nil {
		klog.Warningf("parse pod(%s.%s) annotations[%s] value(%s) failed: %s", pod.GetNamespace(), pod.GetName(), hashKey,
			annotations[hashKey], err.Error())
		return nil
	}
	return containerProbeStatus
}

func GetPodBindProbeAssistantName(pod metav1.Object) (name, namespace string) {
	annotations := pod.GetAnnotations()
	hashKey := ProbeAssistantNameAnnotation
	if annotations[hashKey] == "" {
		return "", ""
	}
	namespaceAndName := strings.Split(annotations[hashKey], "/")
	switch len(namespaceAndName) {
	case 1:
		name = namespaceAndName[0]
		namespace = pod.GetNamespace()
	case 2:
		namespace = namespaceAndName[0]
		name = namespaceAndName[1]
	default:
		klog.Warningf("parse pod(%s.%s) annotations[%s] value(%s).", pod.GetNamespace(), pod.GetName(), hashKey,
			annotations[hashKey])
		return "", ""
	}
	return namespace, name
}

func CleanPodAnnotation(ctx context.Context, cli client.Client, pod *corev1.Pod, pa *appsv1alpha1.ProbeAssistant) error {
	podClone := pod.DeepCopy()
	err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		if value, ok := podClone.Annotations[ProbeAssistantNameAnnotation]; ok && value != "" {
			delete(podClone.Annotations, ProbeAssistantNameAnnotation)
		}
		if value, ok := podClone.Annotations[ProbeAssistantNameChangePodPolicyAnnotation]; ok && value != "" {
			delete(podClone.Annotations, ProbeAssistantNameChangePodPolicyAnnotation)
		}
		if value, ok := podClone.Annotations[ProbeAssistantContainerStatusMapAnnotation]; ok && value != "" {
			delete(podClone.Annotations, ProbeAssistantContainerStatusMapAnnotation)
		}
		if reflect.DeepEqual(pod, podClone) {
			return nil
		}
		return cli.Update(ctx, podClone)
	})
	return err
}

func GetBindingContainers(pod *corev1.Pod) *[]string {
	var containersName []string
	if rawNames, ok := pod.Annotations[ProbeAssistantBindingContainersAnnotation]; ok {
		containersName = strings.Split(rawNames, ",")
	}
	return &containersName
}

func VerifyConatainerNamesOfPod(pod *corev1.Pod, containersName *[]string) error {
	return nil
}

func CheckHasContainerName(name string, containerNames *[]string) bool {
	for _, n := range *containerNames {
		if name == n {
			return true
		}
	}
	return false
}

func MakeProbeAssistantBingdingPod(pod *corev1.Pod, pa *appsv1alpha1.ProbeAssistant) error {
	var containersName *[]string
	containersName = GetBindingContainers(pod)

	if len(*containersName) < 0 {
		klog.V(3).Infof(
			"ProbeAssistant: Pod(%s.%s) did not match any container name, please check pod's annotation(%s).",
			pod.Namespace,
			pod.Name,
			ProbeAssistantBindingContainersAnnotation)
		return nil
	}
	statusMap := make(ProbeAssistantContainerStatusMap)
	for _, container := range pod.Spec.Containers {
		if !CheckHasContainerName(container.Name, containersName) {
			break
		}
		statusMap[pa.Name] = ProbeAssistantContainerStatus{
			pa.Name,
			"",
			"",
			metav1.Now(),
		}
	}
	if statusMapRaw, err := json.Marshal(statusMap); err != nil {
		return err
	} else {
		pod.Annotations[ProbeAssistantContainerStatusMapAnnotation] = string(statusMapRaw)
		pod.Annotations[ProbeAssistantNameAnnotation] = pa.Namespace + "." + pa.Name
	}
	return nil
}
