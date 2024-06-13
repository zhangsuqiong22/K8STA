package controller

import (
	"os"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func newClusterRoleforCRDCreator(name string) *rbacv1.ClusterRole {
	clusterRole := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{"apiextensions.k8s.io"},
				Resources: []string{"customresourcedefinitions"},
				Verbs:     []string{"create", "get", "update", "patch", "delete"},
			},
		},
	}
	return clusterRole
}

func newClusterRoleBindingforCRDCreator(name string, namespace string) *rbacv1.ClusterRoleBinding {
	clusterRoleBinding := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "default",
				Namespace: namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind: "ClusterRole",
			Name: name,
		},
	}
	return clusterRoleBinding
}

func createQtReporterService(name string) *corev1.Service {
	// Define the service object
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": name,
			},
			Ports: []corev1.ServicePort{
				{
					Protocol:   "TCP",
					Port:       8189,
					TargetPort: intstr.FromInt(8189),
				},
			},
		},
	}
	return service
}

func createQtReporterDeployment(name string, namespace string) *appsv1.Deployment {
	image := os.Getenv("QT_REPORTER_IMAGE")
	if image == "" {
		image = "container-infra-local.hzisoj70.china.nsn-net.net/cnfmark/testreporter:v0.2"
	}
	//fmt.Println(image)
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Image:           image,
							ImagePullPolicy: corev1.PullAlways,
							Name:            "crhttpserver",
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("500m"),
									corev1.ResourceMemory: resource.MustParse("1Gi"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("250m"),
									corev1.ResourceMemory: resource.MustParse("512Mi"),
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "kube-config",
									MountPath: "/root/.kube/config",
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  "CRD_NAMESPACE",
									Value: namespace,
								},
							},
							SecurityContext: &corev1.SecurityContext{
								Privileged: ptrBool(true),
								RunAsUser:  int64Ptr(0),
								RunAsGroup: int64Ptr(0),
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "kube-config",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/root/.kube/config",
									Type: new(corev1.HostPathType),
								},
							},
						},
					},
				},
			},
		},
	}
	return deployment
}
