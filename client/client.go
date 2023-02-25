package client

import (
	"context"
	"fmt"
	"os"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type K8s struct {
	*kubernetes.Clientset
}

func NewK8sClient() (*K8s, error) {

	kubeconfig, err := findConfig()
	if err != nil {
		return nil, err
	}

	client, err := kubernetes.NewForConfig(kubeconfig)
	if err != nil {
		return nil, err
	}

	return &K8s{
		client,
	}, nil
}

func findConfig() (*rest.Config, error) {
	var config *rest.Config
	var kubeconfig *string
	var err error

	if os.Getenv("KUBECONFIG") != "" {
		path := os.Getenv("KUBECONFIG")
		kubeconfig = &path
		return buildConfig(kubeconfig)
	}

	home := homedir.HomeDir()
	kubeConfigPath := fmt.Sprintf("%s/%s", home, ".kube/config")

	if _, err := os.Stat(kubeConfigPath); !os.IsNotExist(err) {
		kubeconfig = &kubeConfigPath
		return buildConfig(kubeconfig)

	}

	config, err = rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	return config, nil
}

func buildConfig(kubeconfig *string) (*rest.Config, error) {
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func (k *K8s) ListPods(ctx context.Context, namespace string) (*corev1.PodList, error) {
	podsList, err := k.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})

	return podsList, err
}
