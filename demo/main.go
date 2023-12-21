package main

import (
	"context"
	"encoding/json"
	"fmt"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"os"
)

var clientSet *kubernetes.Clientset
var clientDynamic *dynamic.DynamicClient

func main() {
	initClient()
	initEnvMessage()
	startMenu()
}

func startMenu() {

	for {

		// 读如stdin输入
		fmt.Print("> ")
		var input string
		_, err := fmt.Scanln(&input)
		if err != nil {
			fmt.Println("input error")
		}

		// 以命令行解析
		switch input {
		case "create":
			currentMaxEnv++
			createEnv(currentMaxEnv)
		case "delete":
			var envNum int
			fmt.Print("input envNum: ")
			fmt.Scanln(&envNum)
			destroyEnv(envNum)
		case "list":
			listEnv()
		case "exit":
			os.Exit(0)
		}

	}

}

func initClient() {
	config, err := clientcmd.BuildConfigFromFlags("", "/home/jking/.kube/config")
	if err != nil {
		panic(err)
	}

	clientSet, err = kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	clientDynamic, err = dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}
}

func createEnv(envNum int) {
	//createPV()
	//createPVC()
	var err error
	err = createOSDep(envNum)
	if err != nil {
		fmt.Println(err)
	}
	err = createNgxDep(envNum)
	if err != nil {
		fmt.Println(err)
	}
	err = createOSSvc(envNum)
	if err != nil {
		fmt.Println(err)
	}
	err = createConfigMap(envNum)
	if err != nil {
		fmt.Println(err)
	}
	err = createIngress()
	if err != nil {
		fmt.Println(err)
	}
	//createIngress()
}

func destroyEnv(envNum int) {
	deploymentsClient := clientSet.AppsV1().Deployments(apiv1.NamespaceDefault)
	err := deploymentsClient.Delete(context.TODO(), "ngx-dep-"+fmt.Sprintf("%d", envNum), metav1.DeleteOptions{})
	if err != nil {
		fmt.Println(err)
	}

	err = deploymentsClient.Delete(context.TODO(), "os-dep-"+fmt.Sprintf("%d", envNum), metav1.DeleteOptions{})
	if err != nil {
		fmt.Println(err)
	}

	svcClients := clientSet.CoreV1().Services(apiv1.NamespaceDefault)
	err = svcClients.Delete(context.TODO(), "os-svc-"+fmt.Sprintf("%d", envNum), metav1.DeleteOptions{})
	if err != nil {
		fmt.Println(err)
	}

	configMapClient := clientSet.CoreV1().ConfigMaps(apiv1.NamespaceDefault)
	err = configMapClient.Delete(context.TODO(), "ngx-conf-"+fmt.Sprintf("%d", envNum), metav1.DeleteOptions{})
	if err != nil {
		fmt.Println(err)
	}

	err = deleteIngress(envNum)
	if err != nil {
		fmt.Println(err)
	}
}

func createPV() {

	PVClient := clientSet.CoreV1().PersistentVolumes()

	filename := "pv.json"

	PVConfig := apiv1.PersistentVolume{}

	PVConfigFile, err := os.Open(filename)

	if err != nil {
		panic(err)
	}

	decoder := json.NewDecoder(PVConfigFile)

	err = decoder.Decode(&PVConfig)
	if err != nil {
		panic(err)
	}

	_, err = PVClient.Create(context.TODO(), &PVConfig, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}
}

func createPVC() {

	PVCClient := clientSet.CoreV1().PersistentVolumeClaims(apiv1.NamespaceDefault)

	filename := "pvc.json"

	PVCConfig := apiv1.PersistentVolumeClaim{}

	PVCConfigFile, err := os.Open(filename)

	if err != nil {
		panic(err)
	}

	decoder := json.NewDecoder(PVCConfigFile)

	err = decoder.Decode(&PVCConfig)
	if err != nil {
		panic(err)
	}

	_, err = PVCClient.Create(context.TODO(), &PVCConfig, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}
}

func deleteAll() {
	deploymentsClient := clientSet.AppsV1().Deployments(apiv1.NamespaceDefault)
	err := deploymentsClient.Delete(context.TODO(), "ngx-dep", metav1.DeleteOptions{})
	if err != nil {
		fmt.Println(err)
	}
	err = deploymentsClient.Delete(context.TODO(), "os-dep", metav1.DeleteOptions{})
	if err != nil {
		fmt.Println(err)
	}

	svcClients := clientSet.CoreV1().Services(apiv1.NamespaceDefault)
	err = svcClients.Delete(context.TODO(), "os-svc", metav1.DeleteOptions{})
	if err != nil {
		fmt.Println(err)
	}
	configMapClient := clientSet.CoreV1().ConfigMaps(apiv1.NamespaceDefault)
	err = configMapClient.Delete(context.TODO(), "ngx-conf", metav1.DeleteOptions{})
	if err != nil {
		fmt.Println(err)
	}
	ingressClient := clientSet.NetworkingV1().Ingresses(apiv1.NamespaceDefault)
	err = ingressClient.Delete(context.TODO(), "os-ingress", metav1.DeleteOptions{})
	if err != nil {
		fmt.Println(err)
	}

	PVClient := clientSet.CoreV1().PersistentVolumes()
	err = PVClient.Delete(context.TODO(), "os-pv-volume", metav1.DeleteOptions{})
	if err != nil {
		fmt.Println(err)
	}
	PVCClient := clientSet.CoreV1().PersistentVolumeClaims(apiv1.NamespaceDefault)
	err = PVCClient.Delete(context.TODO(), "os-pv-claim", metav1.DeleteOptions{})
	if err != nil {
		fmt.Println(err)
	}
}
