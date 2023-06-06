package provider

import (
	"github.com/itzloop/iot-vkube/types"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func setPodConditions(status corev1.ConditionStatus, pod *corev1.Pod) (changed bool) {
	// handle PodScheduled
	now := metav1.Now()
	for i, condition := range pod.Status.Conditions {
		if condition.Type != corev1.PodScheduled {
			continue
		}

		if condition.Status == corev1.ConditionTrue {
			//condition.LastProbeTime = now
		} else {
			condition.Status = corev1.ConditionTrue
			//condition.LastProbeTime = now
			condition.LastTransitionTime = now
			condition.Message = "Pod successfully scheduled on the provider"
			condition.Reason = "Scheduled"
		}

		pod.Status.Conditions[i] = condition
		break
	}

	// get condition of type PodReady
	found := false
	for i, condition := range pod.Status.Conditions {
		if condition.Type != corev1.PodReady {
			continue
		}

		// found ready condition
		found = true

		// check if it's changed
		if condition.Status != status {
			// changed
			//condition.LastProbeTime = now
			condition.LastTransitionTime = now
			condition.Status = status
			condition.Reason = "TODO"  // TODO
			condition.Message = "TODO" // TODO
			pod.Status.Conditions[i] = condition
			changed = true
			break
		} else {
			// didn't change
			//condition.LastProbeTime = now
			pod.Status.Conditions[i] = condition
			changed = false
			break
		}
	}

	// if we didn't find ready condition, add it
	if !found {
		changed = true
		pod.Status.Conditions = append(pod.Status.Conditions, corev1.PodCondition{
			Type:   corev1.PodReady,
			Status: status,
			//LastProbeTime: metav1.Now(),
			Reason:  "TODO", //TODO
			Message: "TODO", // TODO
		})
	}

	return
}

func setPodContainerStatuses(pod *corev1.Pod, device types.Device) {
	started := true
	if len(pod.Status.ContainerStatuses) == 0 {
		for _, container := range pod.Spec.Containers {
			pod.Status.ContainerStatuses = append(pod.Status.ContainerStatuses, corev1.ContainerStatus{
				Name: container.Name,
				State: corev1.ContainerState{
					Running: &corev1.ContainerStateRunning{
						StartedAt: metav1.Now(),
					},
				},
				Ready:        device.Ready,
				RestartCount: 0,
				Image:        container.Image,
				Started:      &started,
			})
		}
	} else {
		for _, container := range pod.Spec.Containers {
			for i, cs := range pod.Status.ContainerStatuses {
				if cs.Name != container.Name {
					continue
				}

				// found the container
				// if the status has changed then update it	otherwise ignore it
				if cs.Ready == device.Ready {
					break
				}

				pod.Status.ContainerStatuses[i] = corev1.ContainerStatus{
					Name: container.Name,
					State: corev1.ContainerState{
						Running: &corev1.ContainerStateRunning{
							StartedAt: metav1.Now(),
						},
					},
					Ready:        device.Ready,
					RestartCount: 0,
					Image:        container.Image,
					Started:      &started,
				}
			}
		}
	}
}

func setPodPhase(pod *corev1.Pod, device types.Device) {
	var (
		phase corev1.PodPhase
	)
	if device.Ready {
		phase = corev1.PodRunning
	} else {
		phase = corev1.PodPending
	}

	pod.Status.Message = "TODO: what goes here?"
	pod.Status.Phase = phase
}

func getConditionStatus(pod *corev1.Pod, device types.Device) corev1.ConditionStatus {
	var (
		status corev1.ConditionStatus
	)
	if device.Ready {
		status = corev1.ConditionTrue
	} else {
		status = corev1.ConditionFalse
	}

	return status
}
