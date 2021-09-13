package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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
		ctx                  context.Context
		cancel               context.CancelFunc
		namespace            string
		resourceVersion      string
		pageLimit            int64
		continueToken        string
		resourceVersionMatch string
	)
	timeout := 100 * time.Millisecond
	if err == nil {
		// The request has a timeout, so create a context that is
		// canceled automatically when the timeout expires.
		ctx, cancel = context.WithTimeout(context.Background(), timeout)
	} else {
		ctx, cancel = context.WithCancel(context.Background())
	}
	defer cancel()

	namespace = "deepak" // provide empty string "" for all namespaces
	pageLimit = 5
	continueToken = ""
	resourceVersion = ""
	resourceVersionMatch = ""

	// Get Secrets in a Namespace
	secretList, err := GetSecretsInNamespace(ctx, clientset, namespace, pageLimit, resourceVersion, resourceVersionMatch, continueToken)
	fmt.Printf("There are %d secrets in the namespace %s\n", len(secretList.Items), namespace)
	for _, secret := range secretList.Items {
		fmt.Println(secret.Name)
	}

	continueToken = ""
	resourceVersion = ""
	resourceVersionMatch = ""

	// Get Pods in a Namespace
	podList, err := GetPodsInNamespace(ctx, clientset, namespace, pageLimit, resourceVersion, resourceVersionMatch, continueToken)
	fmt.Printf("There are %d pods in the namespace %s\n", len(podList.Items), namespace)
	for _, pod := range podList.Items {
		fmt.Println(pod.Name)
	}

	continueToken = podList.Continue
	resourceVersion = podList.ResourceVersion
	resourceVersionMatch = "NotOlderThan"
	fmt.Printf("continueToken: %s | resourceVersion: %s\n", continueToken, resourceVersion)

	for continueToken != "" {
		// Get Next set of Pods
		podList, err = GetPodsInNamespace(ctx, clientset, namespace, pageLimit, resourceVersion, resourceVersionMatch, continueToken)
		fmt.Printf("There are %d more pods in the namespace %s\n", len(podList.Items), namespace)
		for _, pod := range podList.Items {
			fmt.Println(pod.Name)
		}

		// https://kubernetes.io/docs/reference/using-api/api-concepts/#resource-versions
		continueToken = podList.Continue
		resourceVersion = podList.ResourceVersion
		fmt.Printf("continueToken: %s | resourceVersion: %s\n", continueToken, resourceVersion)
		time.Sleep(6 * time.Minute)
	}

	// Examples for error handling:
	// - Use helper functions e.g. errors.IsNotFound()
	// - And/or cast to StatusError and use its properties like e.g. ErrStatus.Message
	podName := "auditlog-cleaner-1630119600-tpsbj"
	_, err = GetPodInNamespace(ctx, clientset, namespace, podName)
	if errors.IsNotFound(err) {
		fmt.Printf("Pod %s not found in namespace %s\n", podName, namespace)
	} else if statusError, isStatus := err.(*errors.StatusError); isStatus {
		fmt.Printf("Error getting pod %v\n", statusError.ErrStatus.Message)
	} else if err != nil {
		panic(err.Error())
	} else {
		fmt.Printf("Found %s pod in %s namespace\n", podName, namespace)
	}
}

func GetPodInNamespace(ctx context.Context, clientset *kubernetes.Clientset, namespace string, podName string) (pod *corev1.Pod, err error) {
	// Get Pods in a Namespace
	// get pods in all the namespaces by omitting namespace
	// Or specify namespace to get pods in particular namespace
	return clientset.CoreV1().Pods(namespace).Get(context.TODO(), podName, metav1.GetOptions{})
}

func GetPodsInNamespace(ctx context.Context, clientset *kubernetes.Clientset, namespace string, pageLimit int64, resourceVersion, resourceVersionMatch, continueToken string) (secretList *corev1.PodList, err error) {
	// Get Pods in a Namespace
	// get pods in all the namespaces by omitting namespace
	// Or specify namespace to get pods in particular namespace

	// labelSelector := metav1.LabelSelector{MatchLabels: map[string]string{"version": version}}

	listOptions := metav1.ListOptions{
		// LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
		Limit:    pageLimit,
		Continue: continueToken,
		// ResourceVersion: resourceVersion,
		// ResourceVersionMatch: metav1.ResourceVersionMatch(resourceVersionMatch),
	}

	return clientset.CoreV1().Pods(namespace).List(context.TODO(), listOptions)

}

func GetSecretsInNamespace(ctx context.Context, clientset *kubernetes.Clientset, namespace string, pageLimit int64, resourceVersion, resourceVersionMatch, continueToken string) (secretList *corev1.SecretList, err error) {
	// labelSelector := metav1.LabelSelector{MatchLabels: map[string]string{"version": version}}

	listOptions := metav1.ListOptions{
		// LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
		Limit:    pageLimit,
		Continue: continueToken,
		// ResourceVersion: resourceVersion,
		// ResourceVersionMatch: metav1.ResourceVersionMatch(resourceVersionMatch),
	}
	return clientset.CoreV1().Secrets(namespace).List(ctx, listOptions)
}
