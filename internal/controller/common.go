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

// func createNetworkAttachemntDefinition(name string, namespace string, devicePool string, nadConfig map[string]interface{}) *cniv1.NetworkAttachmentDefinition {
// 	nadConfigJson, err := json.Marshal(nadConfig)
// 	if err != nil {
// 		return nil
// 	}
// 	nadConfigJsonStr := string(nadConfigJson)

// 	// Create new NAD in new namespace
// 	nadObject := &cniv1.NetworkAttachmentDefinition{
// 		ObjectMeta: metav1.ObjectMeta{
// 			Name:      name,
// 			Namespace: namespace,
// 			Annotations: map[string]string{
// 				"k8s.v1.cni.cncf.io/resourceName": devicePool,
// 			},
// 		},
// 		Spec: cniv1.NetworkAttachmentDefinitionSpec{
// 			Config: nadConfigJsonStr,
// 		},
// 	}

// 	return nadObject
// }

func createPodSecurityContext() *corev1.PodSecurityContext {
	return &corev1.PodSecurityContext{
		FSGroup:    int64Ptr(9999),
		RunAsGroup: int64Ptr(9999),
		RunAsUser:  int64Ptr(9999),
	}
}
