# Overview
Kubernetes Test Automation Operator (K8STA) is targeted to execute testing automation cases. It provides below features:
- Easily Trigger and Execute Tests: trigger tests both in and outside of the CI/CD process.
- Retrieve Test Results Efficiently: quickly access test results and artifacts for effective debugging.
- Scalability: efficiently execute parallel test cases by utilizing Kubernetes' native scaling capabilities.
- Support Subset Testing: enabling the execution of specific test cases instead of full testing pipelines.


# Prerequisites
Use the following tools to set up the project:

- Version 1.24 or higher of Go  
- Version 5.3.0 of Kustomize  
- Version 3.14.0 of Kubebuilder  
- The latest version of Docker  
- The latest version of Kubernetes


## Getting Started
You need a in Kubernetes cluster.
**Note:** Your controller will automatically use the current context
in your kubeconfig file (i.e. whatever cluster `kubectl cluster-info` shows).

1. Install test CRDs into cluster:

```sh
$ make install
```

2. Install test controller (switch to a new terminal if you want to leave it running):

```sh
$ make run
```
**NOTE:** You can also run this in one step by running: `make install run`

3. Trigger test cases by test Custom Resources:

```sh
$ kubectl apply -f tester_cr.yaml
```
**NOTE:** package your automation testing cases to container image, replace it to CR in filed: testPodSpec

### Recover Env.
Uninstall CRDs: delete the CRDs from the cluster:

```sh
$ make uninstall
```

### Undeploy controller
UnDeploy the controller to the cluster:

```sh
$ make undeploy
```

### Running on the cluster with PoD (Optional)
1. Build and push your image to the location specified by `IMG`:

```sh
$ make docker-build docker-push IMG=<some-registry>/k8sta:tag
```

2. Deploy the controller to the cluster with the image specified by `IMG`:

```sh
$ make deploy IMG=<some-registry>/k8sta:tag
```

3. Install Instances of Custom Resources(test CR):

```sh
$ kubectl apply -f tester_cr.yaml
```

4. Uninstall Instances of K8STA:

```sh
$ kubectl delete -f kubeta_deployment.yaml
```

## Contributing
// TODO(user): Add detailed information on how you would like others to contribute to this project

### How it works
This project aims to follow the Kubernetes [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/)

It uses [Controllers](https://kubernetes.io/docs/concepts/architecture/controller/)
which provides a reconcile function responsible for synchronizing resources untile the desired state is reached on the cluster

### Modifying the API definitions
If you are editing the API definitions, generate the manifests such as CRs or CRDs using:

```sh
$ make manifests
```

**NOTE:** Run `make --help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)
