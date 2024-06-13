package controller

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newTestStatefulSet(name string, namespace string, image string, replicas int32) *appsv1.StatefulSet {
	statefulsetName := fmt.Sprintf("%s-statefulset", name)
	var argCommand string
	argCommand = "/root/run_case.sh " + " k8s-sno " + " tester-testing-report-manager " + " kubeta "
	statefulset := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      "test-pod",
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": statefulsetName,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: namespace,
					Labels: map[string]string{
						"app": statefulsetName,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            "dpdk",
							Image:           image,
							ImagePullPolicy: corev1.PullAlways,
							Command:         []string{"/bin/bash"},
							Args:            []string{"-c", argCommand},
							Env: []corev1.EnvVar{
								{
									Name: "NODE_NAME",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "spec.nodeName",
										},
									},
								},
							},
						},
					},
					//SecurityContext:               createPodSecurityContext(),
					//RuntimeClassName:              strPtr("nokia-performance"),
					TerminationGracePeriodSeconds: int64Ptr(5),
					Volumes:                       createTestVolumes(),
					Affinity: &corev1.Affinity{
						PodAntiAffinity: &corev1.PodAntiAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
								{
									LabelSelector: &metav1.LabelSelector{
										MatchLabels: map[string]string{
											"app": statefulsetName,
										},
									},
									TopologyKey: "kubernetes.io/hostname",
								},
							},
						},
					},
				},
			},
		},
	}
	return statefulset
}

func createTestVolumes() []corev1.Volume {
	volumes := []corev1.Volume{
		{
			Name: "robot-logs",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
	}

	return volumes
}
