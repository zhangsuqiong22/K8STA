package controller

import (
	//"encoding/json"
	"strings"

	corev1 "k8s.io/api/core/v1"
	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func int64Ptr(i int64) *int64 {
	return &i
}

func int32Ptr(i int32) *int32 {
	return &i
}

func ptrBool(b bool) *bool {
	return &b
}

func strPtr(s string) *string {
	return &s
}

func CheckKeyValuePairExistence(key string, expectedValue string, m map[string]string) bool {
	value, exists := m[key]
	return exists && value == expectedValue
}

func getFirstRunningPodNamesWithPrefix(pods []corev1.Pod, prefix string) (string, error) {
	for _, pod := range pods {
		if pod.GetObjectMeta().GetDeletionTimestamp() != nil {
			continue
		}
		if pod.Status.Phase == corev1.PodPending || pod.Status.Phase == corev1.PodRunning {
			if strings.HasPrefix(pod.Name, prefix) {
				return pod.Name, nil
			}
		}
	}
	return "", nil
}

func createPodSecurityContext() *corev1.PodSecurityContext {
	return &corev1.PodSecurityContext{
		FSGroup:    int64Ptr(9999),
		RunAsGroup: int64Ptr(9999),
		RunAsUser:  int64Ptr(9999),
	}
}
