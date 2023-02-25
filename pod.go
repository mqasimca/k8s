package main

import (
	"context"
	"strings"
	"time"

	"github.com/mqasimca/k8s/client"
	"github.com/spf13/viper"
	corev1 "k8s.io/api/core/v1"
)

func listPods(ctx context.Context, kC *client.K8s, ch chan Pod) error {
	defer close(ch)

	pods, err := kC.ListPods(ctx, "")
	if err != nil {
		return err
	}

	for _, p := range pods.Items {
		counterIncremental["pods"]["total"]++

		if strings.Contains(p.Namespace, "kube-") {
			continue
		}

		podStatusReason := p.Status.Reason
		cStatus := p.Status.ContainerStatuses
		createdAt := createdAt(p.GetCreationTimestamp().UTC().Add(0 * time.Hour))

		pod := Pod{
			PodName:   p.Name,
			Namespace: p.Namespace,
			Reason:    podStatusReason,
			State:     string(p.Status.Phase),
		}

		switch p.Status.Phase {
		case "Succeeded":
			continue

		case "Failed":
			if createdAt > viper.GetFloat64("CREATED_AT") {
				ch <- pod
			}
		case "Pending":
			if createdAt > viper.GetFloat64("CREATED_AT") {
				_, r := getPodConditionFromList(p.Status.Conditions, corev1.PodScheduled)
				pod.Reason = r.Reason
				for _, s := range cStatus {
					if s.State.Waiting != nil {
						if s.State.Waiting.Reason != "" {
							pod.Reason = s.State.Waiting.Reason
						}
					}
				}
				ch <- pod
			}
		case "Running":
			if createdAt > viper.GetFloat64("CREATED_AT") {
				if !isPodReady(&p) {
					for _, s := range cStatus {
						if s.State.Waiting != nil {
							if s.State.Waiting.Reason != "" {
								pod.Reason = s.State.Waiting.Reason
							}
						}
					}
					ch <- pod
				}
			}
		default:
			continue
		}

	}
	return nil
}

// IsPodReady returns true if a pod is ready; false otherwise.
func isPodReady(pod *corev1.Pod) bool {
	return isPodReadyConditionTrue(pod.Status)
}

// IsPodReadyConditionTrue returns true if a pod is ready; false otherwise.
func isPodReadyConditionTrue(status corev1.PodStatus) bool {
	condition := getPodReadyCondition(status)
	return condition != nil && condition.Status == corev1.ConditionTrue
}

// GetPodReadyCondition extracts the pod ready condition from the given status and returns that.
// Returns nil if the condition is not present.
func getPodReadyCondition(status corev1.PodStatus) *corev1.PodCondition {
	_, condition := getPodCondition(&status, corev1.PodReady)
	return condition
}

// GetPodCondition extracts the provided condition from the given status and returns that.
// Returns nil and -1 if the condition is not present, and the index of the located condition.
func getPodCondition(status *corev1.PodStatus, conditionType corev1.PodConditionType) (int, *corev1.PodCondition) {
	if status == nil {
		return -1, nil
	}
	return getPodConditionFromList(status.Conditions, conditionType)
}

// GetPodConditionFromList extracts the provided condition from the given list of condition and
// returns the index of the condition and the condition. Returns -1 and nil if the condition is not present.
func getPodConditionFromList(conditions []corev1.PodCondition, conditionType corev1.PodConditionType) (int, *corev1.PodCondition) {
	if conditions == nil {
		return -1, nil
	}
	for i := range conditions {
		if conditions[i].Type == conditionType {
			return i, &conditions[i]
		}
	}
	return -1, nil
}

// func maxContainerRestarts(pod *corev1.Pod) int {
//  maxRestarts := 0
//  for _, c := range pod.Status.ContainerStatuses {
//      maxRestarts = integer.IntMax(maxRestarts, int(c.RestartCount))
//  }
//  return maxRestarts
// }
