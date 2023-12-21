package main

import (
	"context"
	"fmt"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"testing"
)

func TestOne(t *testing.T) {
	config, err := clientcmd.BuildConfigFromFlags("", "/home/jking/.kube/config")
	if err != nil {
		panic(err)
	}

	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	ingress, err := clientSet.NetworkingV1().Ingresses(apiv1.NamespaceDefault).Get(context.Background(), "os-ingress", metav1.GetOptions{})
	if err != nil {
		panic(err)
	}
	fmt.Println(ingress)

}

func TestYAML2JSON(t *testing.T) {
	var input string
	fmt.Scanln(&input)
	fmt.Println(input)

}
