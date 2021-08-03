package probeassistant

import (
	"context"
	appsv1alpha1 "github.com/twink7e/probeassistant/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	LIVENESS_CONFIGMAP_MOUNT_PATH_PREFIX  = "/probeassistant/liveness"
	READINESS_CONFIGMAP_MOUNT_PATH_PREFIX = "/probeassistant/readiness"
)

func NewConfigMapVolume(confgMapName string) corev1.Volume {
	return corev1.Volume{
		Name: confgMapName,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: confgMapName,
				},
			},
		},
	}
}

func NewConfigMapMount(configMapName string, mountPath string) corev1.VolumeMount {
	return corev1.VolumeMount{
		Name:      configMapName,
		ReadOnly:  true,
		MountPath: mountPath,
	}
}

func InjectMount(ctx context.Context, cli client.Client, pa *appsv1alpha1.ProbeAssistant, pod *corev1.Pod, matchedContainerIndex *[]int) error {
	livenessConfigmapVolue := NewConfigMapVolume(pa.Spec.DefaultLivenessTmpl)
	readinessConfigmapVolue := NewConfigMapVolume(pa.Spec.DefaultReadinessTmpl)
	shouldMountLivenessConfigmap := false
	shouldMountReadinessConfigmap := false
	for _, cid := range *matchedContainerIndex {
		container := &pod.Spec.Containers[cid]

		if container.LivenessProbe != nil {
			shouldMountLivenessConfigmap = true
			container.VolumeMounts = append(container.VolumeMounts, NewConfigMapMount(pa.Spec.DefaultLivenessTmpl, LIVENESS_CONFIGMAP_MOUNT_PATH_PREFIX))
		}
		if container.ReadinessProbe != nil {
			shouldMountReadinessConfigmap = true
			container.VolumeMounts = append(container.VolumeMounts, NewConfigMapMount(pa.Spec.DefaultReadinessTmpl, READINESS_CONFIGMAP_MOUNT_PATH_PREFIX))
		}
	}

	if shouldMountLivenessConfigmap {
		pod.Spec.Volumes = append(pod.Spec.Volumes, livenessConfigmapVolue)
	}
	if shouldMountReadinessConfigmap && pa.Spec.DefaultReadinessTmpl != pa.Spec.DefaultLivenessTmpl {
		pod.Spec.Volumes = append(pod.Spec.Volumes, readinessConfigmapVolue)
	}
	return nil
}
