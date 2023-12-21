package main

import (
	"bufio"
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"os"
	"strconv"
	"strings"
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
		// stdin输入
		// fmt.Scanln seems having some problem
		// I can't read two variable in one line
		// use reader instead
		fmt.Print("> ")
		reader := bufio.NewReader(os.Stdin)
		bytes, _, _ := reader.ReadLine()

		input := strings.Split(string(bytes), " ")

		// 以命令行解析
		switch input[0] {
		case "create":
			createEnv()
		case "delete":
			envNum, err := strconv.Atoi(input[1])
			if err != nil {
				fmt.Println("delete num input error")
				continue
			}

			exist, err := isExistEnv(envNum)
			if err != nil {
				fmt.Println("get env list error")
				continue
			}

			if !exist {
				fmt.Println("env not exist")
				continue
			}

			destroyEnv(envNum)
		case "list":
			listEnv()
		case "exit":
			os.Exit(0)
		}

	}

}

func isExistEnv(envNum int) (bool, error) {
	envNums, err := getAllEnv()
	if err != nil {
		return false, err
	}

	exist := false

	for _, num := range envNums {
		if num == envNum {
			exist = true
			break
		}
	}

	return exist, nil
}

func createEnv() {
	//createPV()
	//createPVC()
	currentMaxEnv++
	envNum := currentMaxEnv
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
	err := deploymentsClient.Delete(context.TODO(),
		fmt.Sprintf("%s%d", ngxDepPrefix, envNum),
		metav1.DeleteOptions{})
	if err != nil {
		fmt.Println(err)
	}

	err = deploymentsClient.Delete(context.TODO(),
		fmt.Sprintf("%s%d", osDepPrefix, envNum),
		metav1.DeleteOptions{})
	if err != nil {
		fmt.Println(err)
	}

	err = svcClient.Delete(context.TODO(),
		fmt.Sprintf("%s%d", osSvcPrefix, envNum),
		metav1.DeleteOptions{})
	if err != nil {
		fmt.Println(err)
	}

	err = configMapClient.Delete(context.TODO(),
		fmt.Sprintf("%s%d", ngxCmPrefix, envNum),
		metav1.DeleteOptions{})
	if err != nil {
		fmt.Println(err)
	}

	err = deleteIngress(envNum)
	if err != nil {
		fmt.Println(err)
	}
}

//func createPV() {
//
//	PVClient := clientSet.CoreV1().PersistentVolumes()
//
//	filename := "pv.json"
//
//	PVConfig := v1.PersistentVolume{}
//
//	PVConfigFile, err := os.Open(filename)
//
//	if err != nil {
//		panic(err)
//	}
//
//	decoder := json.NewDecoder(PVConfigFile)
//
//	err = decoder.Decode(&PVConfig)
//	if err != nil {
//		panic(err)
//	}
//
//	_, err = PVClient.Create(context.TODO(), &PVConfig, metav1.CreateOptions{})
//	if err != nil {
//		panic(err)
//	}
//}
//
//func createPVC() {
//
//	PVCClient := clientSet.CoreV1().PersistentVolumeClaims(v1.NamespaceDefault)
//
//	filename := "pvc.json"
//
//	PVCConfig := v1.PersistentVolumeClaim{}
//
//	PVCConfigFile, err := os.Open(filename)
//
//	if err != nil {
//		panic(err)
//	}
//
//	decoder := json.NewDecoder(PVCConfigFile)
//
//	err = decoder.Decode(&PVCConfig)
//	if err != nil {
//		panic(err)
//	}
//
//	_, err = PVCClient.Create(context.TODO(), &PVCConfig, metav1.CreateOptions{})
//	if err != nil {
//		panic(err)
//	}
//}
