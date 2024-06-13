package controller

import (
	"fmt"
	mytesterv1 "kubeta.github.io/mytester/api/v1"

	"os"
	"strconv"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newRoleforCaseController(namespace string) *rbacv1.Role {
	role := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      "robot-runner",
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{"apps"},
				Resources: []string{"deployments"},
				Verbs:     []string{"create", "delete", "deletecollection", "get", "list", "patch", "update", "watch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"pods", "serviceaccounts", "pods/log", "pods/exec", "persistentvolumeclaims"},
				Verbs:     []string{"create", "delete", "deletecollection", "get", "list", "patch", "update", "watch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"services", "services/proxy"},
				Verbs:     []string{"create", "delete", "get", "list", "patch", "update", "watch"},
			},
			{
				APIGroups:     []string{""},
				Resources:     []string{"secrets"},
				ResourceNames: []string{"singleuser-image-credentials"},
				Verbs:         []string{"list", "watch", "create", "get"},
			},
		},
	}
	return role
}

func newRoleBindingforCaseController(namespace string) *rbacv1.RoleBinding {
	roleBinding := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      "robot-runner",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "default",
				Namespace: namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "Role",
			Name:     "robot-runner",
			APIGroup: "rbac.authorization.k8s.io",
		},
	}
	return roleBinding
}

func newCaseControllerDeployment(name string, namespace string, image string, nodeName string, count int, scope mytesterv1.TestCaseScope) *appsv1.Deployment {
	resources := &corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("1"),
			corev1.ResourceMemory: resource.MustParse("1Gi"),
		},
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("1"),
			corev1.ResourceMemory: resource.MustParse("1Gi"),
		},
	}

	deploymentName := fmt.Sprintf("%s-deployment", name)
	hostIP := os.Getenv("HOST_IP")
	hostUserName := os.Getenv("HOST_USERNAME")
	hostPassword := os.Getenv("HOST_PSW")
	var argCommand string
	nodeName = "k8s-sno"
	if len(hostIP) == 0 {
		argCommand = "/root/run_case.sh " + nodeName
	} else {
		argCommand = "/robot/run_node.sh"
	}
	argCommand = argCommand + " tester-testing-report-manager " + " kubeta " + strconv.Itoa(count)
	if scope.Robot == true {
		argCommand = argCommand + " kubeta-robot-case"
	}
	if scope.Postman == true {
		argCommand = argCommand + " kubeta-postman-case"
	}
	if scope.Cypress == true {
		argCommand = argCommand + " kubeta-cypress-case"
	}
	if scope.Performance == true {
		argCommand = argCommand + " infra-qt-perf"
	}
	if scope.Realtime == true {
		argCommand = argCommand + "  infra-qt-rt"
	}
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "case-controller",
			Namespace: namespace,
		},

		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": deploymentName,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: namespace,
					Labels: map[string]string{
						"app": deploymentName,
					},
					Annotations: map[string]string{
						"k8s.v1.cni.cncf.io/networks": "infra-qt-sp@test1",
					},
				},
				Spec: corev1.PodSpec{
					//SecurityContext: createCaseControllerPodSecurityContext(),
					Containers: []corev1.Container{
						{
							Name:            "robot-container",
							Image:           image,
							ImagePullPolicy: corev1.PullAlways,
							Command:         []string{"/bin/bash"},
							Args:            []string{"-c", argCommand},
							Env: []corev1.EnvVar{
								{
									Name: "MY_POD_NAME",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.name",
										},
									},
								},
								{
									Name:  "MY_POD_NAMESPACE",
									Value: namespace,
								},
								{
									Name:  "MY_CR_NAME",
									Value: name,
								},
								{
									Name:  "FINISH_ANNO_KEY",
									Value: "rcp.nokia.com/" + name + "-status",
								},
								{
									Name:  "FINISH_ANNO_VALUE",
									Value: "finished",
								},
								{
									Name:  "PEER_TEST_POD_NAME",
									Value: "test-peer-pod-0",
								},
								{
									Name:  "HOST_IP",
									Value: hostIP,
								},
								{
									Name:  "HOST_USERNAME",
									Value: hostUserName,
								},
								{
									Name:  "HOST_PASSWORD",
									Value: hostPassword,
								},
							},
							Resources:       *resources,
							SecurityContext: createCaseControllerContainerSecurityContext(),
							VolumeMounts:    createCaseControllerVolumeMounts(),
							Stdin:           true,
							TTY:             true,
						},
					},
					Volumes: createCaseControllerVolumes(),
				},
			},
		},
	}
	return deployment
}

func createCaseControllerContainerSecurityContext() *corev1.SecurityContext {
	securityContext := &corev1.SecurityContext{
		Privileged:             ptrBool(true),
		ReadOnlyRootFilesystem: ptrBool(true),
		SELinuxOptions: &corev1.SELinuxOptions{
			Type: "spc_t",
		},
	}
	return securityContext
}

func createCaseControllerPodSecurityContext() *corev1.PodSecurityContext {
	podSecurityContext := &corev1.PodSecurityContext{
		Sysctls: []corev1.Sysctl{
			{
				Name:  "net.ipv6.conf.all.disable_ipv6",
				Value: "0",
			},
			{
				Name:  "net.ipv4.ip_local_port_range",
				Value: "54000 65535",
			},
			{
				Name:  "net.ipv6.conf.all.accept_ra",
				Value: "0",
			},
			{
				Name:  "net.ipv4.conf.all.rp_filter",
				Value: "0",
			},
			{
				Name:  "net.ipv4.conf.default.rp_filter",
				Value: "0",
			},
			{
				Name:  "net.ipv4.conf.default.disable_policy",
				Value: "1",
			},
			{
				Name:  "net.ipv4.conf.all.disable_policy",
				Value: "1",
			},
			{
				Name:  "net.ipv6.conf.default.disable_policy",
				Value: "1",
			},
			{
				Name:  "net.ipv6.conf.all.disable_policy",
				Value: "1",
			},
		},
	}
	return podSecurityContext
}

func createCaseControllerVolumeMounts() []corev1.VolumeMount {
	volumenMounts := []corev1.VolumeMount{
		{
			Name:      "case-logs",
			MountPath: "/case/logs",
			ReadOnly:  false,
		},
		{
			Name:      "tmp",
			MountPath: "/tmp",
			ReadOnly:  false,
		},
	}

	return volumenMounts
}

func createCaseControllerVolumes() []corev1.Volume {
	volumes := []corev1.Volume{
		{
			Name: "shm",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{
					Medium: "Memory",
				},
			},
		},
		{
			Name: "case-logs",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
		{
			Name: "tmp",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
	}

	return volumes
}
