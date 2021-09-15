package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	// // creates the in-cluster config
	// config, err := rest.InClusterConfig()
	// if err != nil {
	// 	panic(err.Error())
	// }

	kubeconfig := os.Getenv("KUBECONFIG")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Panic(err.Error())
	}

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Panic(err.Error())
	}

	var (
		ctx       context.Context
		cancel    context.CancelFunc
		namespace string
	)

	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()

	namespace = "deepak" // provide empty string "" for all namespaces

	PrintSecretsPaginated(ctx, clientset, namespace)
	PrintDeploymentsPaginated(ctx, clientset, namespace)
	var deploymentLabels map[string]string
	PrintPodsPaginated(ctx, clientset, namespace, deploymentLabels)

	deploymentName := "auditlog-deployment"
	deployment, err := GetDeploymentInNamespace(ctx, clientset, namespace, deploymentName)
	if errors.IsNotFound(err) {
		fmt.Printf("Deployment %s not found in namespace %s\n", deploymentName, namespace)
	} else if statusError, isStatus := err.(*errors.StatusError); isStatus {
		fmt.Printf("Error getting deployment %v\n", statusError.ErrStatus.Message)
	} else if err != nil {
		panic(err.Error())
	} else {
		fmt.Printf("Found %s deployment in %s namespace\n", deploymentName, namespace)
	}
	deploymentJSON, err := json.MarshalIndent(deployment, "", "  ")
	_ = deploymentJSON
	if err != nil {
		log.Fatalf(err.Error())
	}
	fmt.Printf("Deployment Labels: %s\n", deployment.GetLabels())
	//var deploymentLabels map[string]string
	deploymentLabels = make(map[string]string)
	deploymentLabels["app"] = "auditlog"
	PrintPodsPaginated(ctx, clientset, namespace, deploymentLabels)
	//fmt.Printf("Deployment Details: %s\n", string(deploymentJSON))

	// Examples for error handling:
	// - Use helper functions e.g. errors.IsNotFound()
	// - And/or cast to StatusError and use its properties like e.g. ErrStatus.Message
	podName := "auditlog-deployment-5d4755666d-fgvwd"
	pod, err := GetPodInNamespace(ctx, clientset, namespace, podName)
	if errors.IsNotFound(err) {
		fmt.Printf("Pod %s not found in namespace %s\n", podName, namespace)
	} else if statusError, isStatus := err.(*errors.StatusError); isStatus {
		fmt.Printf("Error getting pod %v\n", statusError.ErrStatus.Message)
	} else if err != nil {
		panic(err.Error())
	} else {
		fmt.Printf("Found %s pod in %s namespace\n", podName, namespace)
	}
	//MarshalIndent
	podJSON, err := json.MarshalIndent(pod, "", "  ")
	_ = podJSON
	if err != nil {
		log.Fatalf(err.Error())
	}
	fmt.Printf("Pod Labels: %s\n", pod.GetLabels())
	//fmt.Printf("Pod Details: %s\n", string(podJSON))
}

func PrintPodsPaginated(ctx context.Context, clientset *kubernetes.Clientset, namespace string, deploymentLabels map[string]string) {
	var resourceVersion string
	var continueToken string
	var resourceVersionMatch string
	var pageLimit int64
	var remainingItemCount int64

	continueToken = ""
	resourceVersion = ""
	resourceVersionMatch = "NotOlderThan"
	pageLimit = 5
	remainingItemCount = 0

	// Get Pods in a Namespace
	podList, _ := GetPodsInNamespace(ctx, clientset, namespace, pageLimit, resourceVersion, resourceVersionMatch, continueToken, deploymentLabels)
	if podList.GetRemainingItemCount() != nil {
		remainingItemCount = *podList.GetRemainingItemCount()
	} else {
		remainingItemCount = 0
	}
	fmt.Printf("There are %d pods in the namespace %s with remaining: %#v\n", len(podList.Items), namespace, remainingItemCount)
	for _, pod := range podList.Items {
		fmt.Printf("[%s] labels=%s\n", pod.GetName(), pod.GetLabels())
	}

	continueToken = podList.Continue
	resourceVersion = podList.ResourceVersion
	fmt.Printf("continueToken: %s | resourceVersion: %s\n", continueToken, resourceVersion)

	for continueToken != "" {
		// Get Next set of Pods
		podList, _ = GetPodsInNamespace(ctx, clientset, namespace, pageLimit, resourceVersion, resourceVersionMatch, continueToken, deploymentLabels)
		if podList.GetRemainingItemCount() != nil {
			remainingItemCount = *podList.GetRemainingItemCount()
		} else {
			remainingItemCount = 0
		}
		fmt.Printf("There are %d more pods in the namespace %s with remaining: %#v\n", len(podList.Items), namespace, remainingItemCount)
		for _, pod := range podList.Items {
			fmt.Printf("[%s] labels=%s\n", pod.GetName(), pod.GetLabels())
		}

		// https://kubernetes.io/docs/reference/using-api/api-concepts/#resource-versions
		continueToken = podList.Continue
		resourceVersion = podList.ResourceVersion
		fmt.Printf("continueToken: %s | resourceVersion: %s\n", continueToken, resourceVersion)
		// time.Sleep(6 * time.Minute)
	}
}

func PrintDeploymentsPaginated(ctx context.Context, clientset *kubernetes.Clientset, namespace string) {
	var resourceVersion string
	var continueToken string
	var resourceVersionMatch string
	var pageLimit int64
	var remainingItemCount int64

	continueToken = ""
	resourceVersion = ""
	resourceVersionMatch = "NotOlderThan"
	pageLimit = 5
	remainingItemCount = 0

	// Get all Deployments
	deploymentList, _ := GetDeploymentsInNamespace(ctx, clientset, namespace, pageLimit, resourceVersion, resourceVersionMatch, continueToken)
	if deploymentList.GetRemainingItemCount() != nil {
		remainingItemCount = *deploymentList.GetRemainingItemCount()
	} else {
		remainingItemCount = 0
	}
	fmt.Printf("There are %d deployments in the namespace %s with remaining: %#v\n", len(deploymentList.Items), namespace, remainingItemCount)
	for _, deployment := range deploymentList.Items {
		fmt.Println(deployment.Name)
	}

	continueToken = deploymentList.Continue
	resourceVersion = deploymentList.ResourceVersion
	fmt.Printf("continueToken: %s | resourceVersion: %s\n", continueToken, resourceVersion)

	for continueToken != "" {
		// Get Next set of Pods
		deploymentList, _ = GetDeploymentsInNamespace(ctx, clientset, namespace, pageLimit, resourceVersion, resourceVersionMatch, continueToken)
		if deploymentList.GetRemainingItemCount() != nil {
			remainingItemCount = *deploymentList.GetRemainingItemCount()
		} else {
			remainingItemCount = 0
		}
		fmt.Printf("There are %d more deployments in the namespace %s with remaining: %#v\n", len(deploymentList.Items), namespace, remainingItemCount)
		for _, deployment := range deploymentList.Items {
			fmt.Println(deployment.Name)
		}

		// https://kubernetes.io/docs/reference/using-api/api-concepts/#resource-versions
		continueToken = deploymentList.Continue
		resourceVersion = deploymentList.ResourceVersion
		fmt.Printf("continueToken: %s | resourceVersion: %s\n", continueToken, resourceVersion)
	}

}

func PrintSecretsPaginated(ctx context.Context, clientset *kubernetes.Clientset, namespace string) {
	var resourceVersion string
	var continueToken string
	var resourceVersionMatch string
	var pageLimit int64

	continueToken = ""
	resourceVersion = ""
	resourceVersionMatch = "NotOlderThan"
	pageLimit = 5

	// Get Secrets in a Namespace
	secretList, _ := GetSecretsInNamespace(ctx, clientset, namespace, pageLimit, resourceVersion, resourceVersionMatch, continueToken)
	fmt.Printf("There are %d secrets in the namespace %s\n", len(secretList.Items), namespace)
	for _, secret := range secretList.Items {
		fmt.Println(secret.Name)
	}
}

func GetPodInNamespace(ctx context.Context, clientset *kubernetes.Clientset, namespace string, podName string) (pod *corev1.Pod, err error) {
	return clientset.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
}

func GetDeploymentInNamespace(ctx context.Context, clientset *kubernetes.Clientset, namespace string, deploymentName string) (*appsv1.Deployment, error) {
	deploymentsClient := clientset.AppsV1().Deployments(namespace)
	return deploymentsClient.Get(ctx, deploymentName, metav1.GetOptions{})
}

func GetPodsInNamespace(ctx context.Context, clientset *kubernetes.Clientset, namespace string,
	pageLimit int64, resourceVersion, resourceVersionMatch, continueToken string, podLabels map[string]string) (*corev1.PodList, error) {
	// get pods in all the namespaces by omitting namespace
	// Or specify namespace to get pods in particular namespace

	var listOptions metav1.ListOptions
	if len(podLabels) > 0 {
		fmt.Println("using label selector")
		labelSelector := metav1.LabelSelector{MatchLabels: podLabels}

		listOptions = metav1.ListOptions{
			LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
			Limit:         pageLimit,
			Continue:      continueToken,
		}
	} else {
		listOptions = metav1.ListOptions{
			Limit:    pageLimit,
			Continue: continueToken,
		}
	}
	return clientset.CoreV1().Pods(namespace).List(ctx, listOptions)

}

func GetDeploymentsInNamespace(ctx context.Context, clientset *kubernetes.Clientset, namespace string,
	pageLimit int64, resourceVersion, resourceVersionMatch, continueToken string) (*appsv1.DeploymentList, error) {
	// labelSelector := metav1.LabelSelector{MatchLabels: map[string]string{"version": version}}

	listOptions := metav1.ListOptions{
		// LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
		Limit:    pageLimit,
		Continue: continueToken,
	}

	deploymentsClient := clientset.AppsV1().Deployments(namespace)

	return deploymentsClient.List(ctx, listOptions)

}

func GetSecretsInNamespace(ctx context.Context, clientset *kubernetes.Clientset, namespace string,
	pageLimit int64, resourceVersion, resourceVersionMatch, continueToken string) (*corev1.SecretList, error) {
	// labelSelector := metav1.LabelSelector{MatchLabels: map[string]string{"version": version}}

	listOptions := metav1.ListOptions{
		// LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
		Limit:    pageLimit,
		Continue: continueToken,
	}
	return clientset.CoreV1().Secrets(namespace).List(ctx, listOptions)
}
