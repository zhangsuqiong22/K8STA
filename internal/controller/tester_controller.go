/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"

	ext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	//apiextv1 "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	//"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	mytesterv1 "kubeta.github.io/mytester/api/v1"
)

// TesterReconciler reconciles a Tester object
type TesterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=mytester.kubeta.github.io,resources=testers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=mytester.kubeta.github.io,resources=testers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=mytester.kubeta.github.io,resources=testers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Tester object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *TesterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	infra := &mytesterv1.Tester{}
	if err := r.Get(ctx, req.NamespacedName, infra); err != nil {
		if errors.IsNotFound(err) {
			l.Info("the kubeta request is not found")
			return ctrl.Result{}, nil
		}
		l.Error(err, "failed to get the kubeta request")
		return ctrl.Result{}, err
	}

	if infra.Status.CaseStatus == "" {
		l.Info("case status is empty, add status")
		if err := r.updateCRStatus(ctx, infra, req.NamespacedName.Name, "Settingup"); err != nil {
			l.Info("Error when updating status to setting up.")
			//panic(err)
			time.Sleep(1 * time.Second)
		}
		l.Info("Update CR status to Setting up")
	}
	if infra.Status.CaseStatus == "Finished" || infra.Status.CaseStatus == "Tearingdown" {
		return ctrl.Result{}, nil
	}
	role := &rbacv1.Role{}
	err := r.Get(ctx, types.NamespacedName{Name: "robot-runner", Namespace: req.NamespacedName.Namespace}, role)
	if err != nil && errors.IsNotFound(err) {
		role := newRoleforCaseController(req.NamespacedName.Namespace)
		if err := r.Create(ctx, role); err != nil {
			return ctrl.Result{}, err
		}
		l.Info("created new role")
	}

	crdList := &ext.CustomResourceDefinitionList{}
	err = r.List(ctx, crdList)
	if err != nil {
		l.Info("Failed to list Pods.")
		panic(err)
	}
	// sccCrdFound := false
	// sccCrdName := "securitycontextconstraints.security.openshift.io"
	// for _, crd := range crdList.Items {
	// 	if crd.ObjectMeta.Name == sccCrdName {
	// 		sccCrdFound = true
	// 		l.Info("found CRD securitycontextconstraints.security.openshift.io")
	// 	}
	// }
	roleBinding := &rbacv1.RoleBinding{}
	err = r.Get(ctx, types.NamespacedName{Name: "robot-runner", Namespace: req.NamespacedName.Namespace}, roleBinding)
	if err != nil && errors.IsNotFound(err) {
		roleBinding := newRoleBindingforCaseController(req.NamespacedName.Namespace)
		if err := r.Create(ctx, roleBinding); err != nil {
			return ctrl.Result{}, err
		}
		l.Info("created new role binding")
	}

	// Fetch the test pods
	// Get the list of nodes
	nodeList := &corev1.NodeList{}
	err = r.List(ctx, nodeList)
	if err != nil {
		return ctrl.Result{}, err
	}
	nodeCount := len(nodeList.Items)
	l.Info("cluster has", strconv.Itoa(nodeCount), "nodes")

	TestSsFound := &appsv1.StatefulSet{}
	err = r.Get(ctx, types.NamespacedName{Name: "test-pod", Namespace: req.NamespacedName.Namespace}, TestSsFound)
	if err != nil && errors.IsNotFound(err) {
		newTestSS := newTestStatefulSet(req.NamespacedName.Name, req.NamespacedName.Namespace, infra.Spec.TestPodSpec.Image, int32(nodeCount))
		if err := r.Create(ctx, newTestSS); err != nil {
			l.Error(err, "failed to create new test statefulset")
			return ctrl.Result{}, err
		}
		l.Info("created new test statefulset", "Name", newTestSS.Name)

		allRunning, _ := r.areAllPodsRunningInNamespace(ctx, req.NamespacedName.Namespace)
		if allRunning == false {
			l.Info("not all Pods status are running, wait ...")
			return ctrl.Result{Requeue: true, RequeueAfter: time.Second * 3}, nil
		}
	}

	// Fetch the Case controller
	CaseDpFound := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: "case-controller", Namespace: req.NamespacedName.Namespace}, CaseDpFound)
	if err != nil && errors.IsNotFound(err) {
		// Get controller pod name
		podName := os.Getenv("HOSTNAME")
		pod := &corev1.Pod{}
		err = r.Get(ctx, types.NamespacedName{Name: podName, Namespace: "tester-operator-system"}, pod)
		nodeName := pod.Spec.NodeName
		l.Info("created new case deployment", "Name", nodeName)
		// Create the Case controller
		newCaseDp := newCaseControllerDeployment(req.NamespacedName.Name, req.NamespacedName.Namespace, infra.Spec.TestPodSpec.Image, nodeName, nodeCount, infra.Spec.TestCaseScope)
		if err := r.Create(ctx, newCaseDp); err != nil {
			l.Error(err, "failed to create new case deployment")
			return ctrl.Result{}, err
		}
		l.Info("created new case deployment", "Name", newCaseDp.Name)
		// update CR status to running
		if err := r.updateCRStatus(ctx, infra, req.NamespacedName.Name, "Running"); err != nil {
			l.Info("Error when updating status to running.")
			panic(err)
		}
		l.Info("Update CR status to Running")
	}

	if finished := r.isCaseFinished(ctx, req.NamespacedName.Namespace, req.NamespacedName.Name); finished == false {
		// finished flag is not set, keep checking
		return ctrl.Result{Requeue: true, RequeueAfter: time.Second * 3}, nil
	}
	// case is finished, update CR status to tearingdown
	if err := r.updateCRStatus(ctx, infra, req.NamespacedName.Name, "Tearingdown"); err != nil {
		l.Info("Error when updating status to tearingdown.")
		panic(err)
	}
	l.Info("Update CR status to Tearing down")

	if infra.Spec.TestDebugMode == true {
		if err := r.updateCRStatus(ctx, infra, req.NamespacedName.Name, "Finished"); err != nil {
			l.Info("Error when updating status to finished.")
			panic(err)
		}
		l.Info("Update CR status to Finished")
		return ctrl.Result{}, nil
	}
	// delete case controller
	l.Info("Try to delete case controller")
	if err := r.Delete(ctx, CaseDpFound); err != nil {
		l.Info("Error to delete case controller")
	}
	// delete rolebinding and role
	caseControllerRoleBinding := &rbacv1.RoleBinding{}
	err = r.Get(ctx, types.NamespacedName{Name: "robot-runner", Namespace: req.NamespacedName.Namespace}, caseControllerRoleBinding)
	if err == nil {
		if err := r.Delete(ctx, caseControllerRoleBinding); err != nil {
			l.Info("Error to delete case controller rolebinding")
		}
	}

	caseControllerRole := &rbacv1.Role{}
	err = r.Get(ctx, types.NamespacedName{Name: "robot-runner", Namespace: req.NamespacedName.Namespace}, caseControllerRole)
	if err == nil {
		if err := r.Delete(ctx, caseControllerRole); err != nil {
			l.Info("Error to delete case controller role")
		}
	}

	// delete test statefulset
	//TestSsFound := &appsv1.StatefulSet{}

	l.Info("Try to delete test statefulset")
	if err := r.Delete(ctx, TestSsFound); err != nil {
		l.Info("Error to delete test statefulset")
	}

	//---
	// set CR status to finished
	if err := r.updateCRStatus(ctx, infra, req.NamespacedName.Name, "Settingup"); err != nil {
		l.Info("Error when updating status to finished.")
		panic(err)
	}
	l.Info("Update CR status to Finished")

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *TesterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	//fmt.Printf("##############Test###################")
	// return ctrl.NewControllerManagedBy(mgr).
	// 	For(&mytesterv1.Tester{}).
	// 	Complete(r)
	if err := ctrl.NewControllerManagedBy(mgr).
		For(&mytesterv1.Tester{}).
		Complete(r); err != nil {
		return err
	}
	// Connnect to kubeconfig
	// config, err := rest.InClusterConfig()
	// //config, err := config.GetConfig()
	// if err != nil {
	// 	return err
	// }

	// kubeconfig path /root/.kube/config
	kubeconfigPath := "/root/.kube/config"

	// clientcmd add
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		fmt.Printf("Error building kubeconfig: %v\n", err)
		os.Exit(1)
	}

	// Create new kubernetes clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}
	// Create a CRD client
	crdClient, err := apiextv1.NewForConfig(config)
	if err != nil {
		return err
	}
	newNs, err := clientset.CoreV1().Namespaces().Get(context.Background(), "kubeta", metav1.GetOptions{})
	if err != nil && errors.IsNotFound(err) {
		// Define the namespace object
		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "kubeta",
			},
		}

		// Create new rolebinding in new namespace
		sccCrdName := "securitycontextconstraints.security.openshift.io"
		crdList, err := crdClient.CustomResourceDefinitions().List(context.Background(), metav1.ListOptions{})
		if err != nil {
			fmt.Printf("failed to get CRD")
			return err
		}

		sccCrdFound := false
		for _, crd := range crdList.Items {
			if crd.ObjectMeta.Name == sccCrdName {
				sccCrdFound = true
				fmt.Printf("found CRD securitycontextconstraints.security.openshift.io")
			}
		}

		if sccCrdFound == true {
			sccName := "system:openshift:scc:privileged"
			sccRoleBinding := r.createSCCRoleBinding(sccName, newNs.Name)
			_, err = clientset.RbacV1().RoleBindings(newNs.Name).Create(context.Background(), sccRoleBinding, metav1.CreateOptions{})
			if err != nil {
				return err
			}
			fmt.Printf("rolebinding is created")
		}
		// Create a new namespace
		newNs, err = clientset.CoreV1().Namespaces().Create(context.Background(), ns, metav1.CreateOptions{})
		if err != nil {
			if statusErr, isStatus := err.(*errors.StatusError); isStatus && statusErr.Status().Code == http.StatusConflict {
				fmt.Printf("namespace is already existed")
			} else {
				return err
			}
		}
		fmt.Printf("namespace is created")
	} else {
		return err
	}

	// Create rolebinding for crd creator
	// Create rolebinding for crd creator
	crdCreatorRoleName := "crd-creator"
	crdCreatorRole := newClusterRoleforCRDCreator(crdCreatorRoleName)
	_, err = clientset.RbacV1().ClusterRoles().Create(context.Background(), crdCreatorRole, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	crdCreatorRoleBinding := newClusterRoleBindingforCRDCreator(crdCreatorRoleName, newNs.Name)
	_, err = clientset.RbacV1().ClusterRoleBindings().Get(context.Background(), crdCreatorRoleBinding.Name, metav1.GetOptions{})
	if err != nil && errors.IsNotFound(err) {
		_, err = clientset.RbacV1().ClusterRoleBindings().Create(context.Background(), crdCreatorRoleBinding, metav1.CreateOptions{})
		if err != nil {
			return err
		}
		fmt.Printf("rolebinding is created")
	} else if err != nil {
		return err
	} else {
		fmt.Printf("rolebinding already exists")
	}

	//Create new Report deployment
	qtReportName := "tester-testing-report-manager"
	qtReporter := createQtReporterDeployment(qtReportName, newNs.Name)
	_, err = clientset.AppsV1().Deployments(newNs.Name).Create(context.Background(), qtReporter, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	fmt.Printf("reporter is created")

	//Create new service
	qtReporterService := createQtReporterService(qtReportName)
	_, err = clientset.CoreV1().Services(newNs.Name).Create(context.Background(), qtReporterService, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	fmt.Printf("service is created")
	//Create case service
	caseSVCName := "case-controller"
	caseSVC := createCaseControllerService(caseSVCName)
	_, err = clientset.CoreV1().Services(newNs.Name).Create(context.Background(), caseSVC, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}

// createCaseControllerService
func createCaseControllerService(name string) *corev1.Service {
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
					Port:       5201,
					TargetPort: intstr.FromInt(5201),
				},
			},
		},
	}
	return service
}

func (r *TesterReconciler) createSCCRoleBinding(name string, namespace string) *rbacv1.RoleBinding {
	roleBinding := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
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
	return roleBinding
}

func (r *TesterReconciler) areAllPodsRunningInNamespace(ctx context.Context, namespace string) (bool, error) {
	// Fetch the Pods in the namespace.
	podList := &corev1.PodList{}
	if err := r.List(ctx, podList, &client.ListOptions{
		Namespace: namespace,
	}); err != nil {
		return false, err
	}

	// Check the status of each Pod.
	for _, pod := range podList.Items {
		if pod.Status.Phase != corev1.PodRunning {
			// If any pod is not in the "Running" state, return false.
			return false, nil
		}
	}

	// If all pods are in the "Running" state, return true.
	return true, nil
}

func (r *TesterReconciler) updateCRStatus(ctx context.Context, infra *mytesterv1.Tester, crName string, newStatus string) error {
	infra.Status.CaseStatus = newStatus
	newCondition := metav1.Condition{
		Type:               newStatus,
		Status:             metav1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Reason:             "ReconciliationSucceeded",
		Message:            "CR " + crName + " is " + newStatus,
	}
	infra.Status.Conditions = append(infra.Status.Conditions, newCondition)
	return r.Status().Update(ctx, infra)
}

func (r *TesterReconciler) isCaseFinished(ctx context.Context, namespace string, name string) bool {
	l := log.FromContext(ctx)

	allRunning, _ := r.areAllPodsRunningInNamespace(ctx, namespace)
	if allRunning == false {
		l.Info("not all Pods status are running, wait ...")
		return false
	}

	CaseDpFound := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{Name: "case-controller", Namespace: namespace}, CaseDpFound)
	if err != nil {
		l.Info("case controller deployment case-controller NOT found")
		return false
	}
	// Get case controller annotation
	podList := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(CaseDpFound.Namespace),
		client.MatchingLabels(CaseDpFound.Labels),
	}
	err = r.List(ctx, podList, listOpts...)
	if err != nil {
		l.Info("Failed to list Pods.", "Deployment.Namespace", CaseDpFound.Namespace, "Deployment.Name", CaseDpFound.Name)
		return false
	}
	podName, _ := getFirstRunningPodNamesWithPrefix(podList.Items, "case-controller")
	if podName == "" {
		l.Info("Not find case controller")
		return false
	}
	casePod := &corev1.Pod{}
	err = r.Get(ctx, types.NamespacedName{Name: podName, Namespace: CaseDpFound.Namespace}, casePod)
	if err != nil && errors.IsNotFound(err) {
		l.Info("Failed to find case pod, shall not be here.")
		return false
	} else if err != nil {
		l.Info("Error when finding case pod, shall not be here.")
		return false
	}

	annotations := casePod.ObjectMeta.Annotations
	key := "mytester.kubeta.github.io/" + name + "-status"
	return CheckKeyValuePairExistence(key, "finished", annotations)
}
