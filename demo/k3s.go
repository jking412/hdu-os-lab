package main

import (
	"context"
	"encoding/json"
	"fmt"
	appv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"strconv"
	"strings"
)

const resourcePath = "/home/jking/GolandProjects/os-online/demo/resource"
const cmName = "cm.json"
const osDepName = "os-dep.json"
const ngxDepName = "ngx-dep.json"
const osSvcName = "os-svc.json"
const ingressName = "os-ingress"

var currentMaxEnv = 0

var ngxConf = `upstream codeserver {
  server %s:4000;
}
server {
  listen 80;
  location /osenv/%d/ {
        proxy_pass http://codeserver/;
      proxy_set_header Host $http_host;
      proxy_set_header Upgrade $http_upgrade;
      proxy_set_header Connection upgrade;
      proxy_set_header Accept-Encoding gzip;
    }

}`

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

	ingressClient = clientSet.NetworkingV1().Ingresses(v1.NamespaceDefault)
	deploymentsClient = clientSet.AppsV1().Deployments(v1.NamespaceDefault)
	svcClient = clientSet.CoreV1().Services(v1.NamespaceDefault)
	configMapClient = clientSet.CoreV1().ConfigMaps(v1.NamespaceDefault)
}

// ingress struct should not be created by Ingress.Get()
// if use the struct, you will fail to create a new ingress
func createIngress() error {

	var tempIngress *networkingv1.Ingress

	oldIngress, err := ingressClient.Get(context.Background(), ingressName, metav1.GetOptions{})
	if err != nil {
		tempIngress = osIngressT.DeepCopy()
	} else {
		tempIngress = new(networkingv1.Ingress)
		tempIngress.Spec = oldIngress.Spec
		tempIngress.Name = ingressName
	}

	pathType := new(networkingv1.PathType)
	*pathType = networkingv1.PathTypePrefix

	tempIngress.Spec.Rules[0].HTTP.Paths = append(tempIngress.Spec.Rules[0].HTTP.Paths, networkingv1.HTTPIngressPath{
		Path:     fmt.Sprintf("%s%d", ingressPathPrefix, currentMaxEnv),
		PathType: pathType,
		Backend: networkingv1.IngressBackend{
			Service: &networkingv1.IngressServiceBackend{
				Name: fmt.Sprintf("%s%d", osSvcPrefix, currentMaxEnv),
				Port: networkingv1.ServiceBackendPort{
					Number: 80,
				},
			},
		},
	})

	// 先删除原来的ingress
	// 不存在就不删除
	if err != nil {
		_, err = ingressClient.Create(context.Background(), tempIngress, metav1.CreateOptions{})
		if err != nil {
			return err
		}
	} else {
		err = ingressClient.Delete(context.Background(), ingressName, metav1.DeleteOptions{})
		if err != nil {
			return err
		}
		_, err = ingressClient.Create(context.Background(), tempIngress, metav1.CreateOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}

func deleteIngress(envNum int) error {

	oldIngress, err := ingressClient.Get(context.Background(), ingressName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	tempIngress := osIngressT.DeepCopy()

	for i, v := range oldIngress.Spec.Rules[0].HTTP.Paths {
		if v.Path == fmt.Sprintf("%s%d", ingressPathPrefix, envNum) {
			if len(oldIngress.Spec.Rules[0].HTTP.Paths) == 1 {
				oldIngress.Spec.Rules[0].HTTP.Paths = make([]networkingv1.HTTPIngressPath, 0)
			} else if i == len(oldIngress.Spec.Rules[0].HTTP.Paths)-1 {
				oldIngress.Spec.Rules[0].HTTP.Paths = oldIngress.Spec.Rules[0].HTTP.Paths[:i]
			} else {
				oldIngress.Spec.Rules[0].HTTP.Paths = append(oldIngress.Spec.Rules[0].HTTP.Paths[:i], oldIngress.Spec.Rules[0].HTTP.Paths[i+1:]...)
			}
			break
		}
	}

	tempIngress.Spec = oldIngress.Spec

	if len(oldIngress.Spec.Rules[0].HTTP.Paths) == 0 {
		err = ingressClient.Delete(context.Background(), ingressName, metav1.DeleteOptions{})
		if err != nil {
			return err
		}
	} else {
		err = ingressClient.Delete(context.Background(), ingressName, metav1.DeleteOptions{})
		if err != nil {
			return err
		}
		_, err = ingressClient.Create(context.Background(), tempIngress, metav1.CreateOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}

func createOSSvc(envNum int) error {

	filename := resourcePath + "/" + osSvcName

	svcConfig := v1.Service{}
	svcConfigFile, err := os.Open(filename)
	if err != nil {
		return err
	}

	decoder := json.NewDecoder(svcConfigFile)

	err = decoder.Decode(&svcConfig)
	if err != nil {
		return err
	}

	svcConfig.Name = svcConfig.Name + "-" + fmt.Sprintf("%d", envNum)
	svcConfig.Spec.Selector["app"] = svcConfig.Spec.Selector["app"] + "-" + fmt.Sprintf("%d", envNum)

	_, err = svcClient.Create(context.TODO(), &svcConfig, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func createNgxDep(envNum int) error {

	deploymentsConfig := appv1.Deployment{}

	filename := resourcePath + "/" + ngxDepName

	deploymentsConfigFile, err := os.Open(filename)
	if err != nil {
		return err
	}

	decoder := json.NewDecoder(deploymentsConfigFile)

	err = decoder.Decode(&deploymentsConfig)
	if err != nil {
		return err
	}

	deploymentsConfig.Name = deploymentsConfig.Name + "-" + fmt.Sprintf("%d", envNum)
	deploymentsConfig.Spec.Selector.MatchLabels["ngx"] = deploymentsConfig.Spec.Selector.MatchLabels["ngx"] + "-" + fmt.Sprintf("%d", envNum)
	deploymentsConfig.Spec.Template.ObjectMeta.Labels["ngx"] = deploymentsConfig.Spec.Template.ObjectMeta.Labels["ngx"] + "-" + fmt.Sprintf("%d", envNum)
	deploymentsConfig.Spec.Template.ObjectMeta.Labels["app"] = deploymentsConfig.Spec.Template.ObjectMeta.Labels["app"] + "-" + fmt.Sprintf("%d", envNum)
	deploymentsConfig.Spec.Template.Spec.Volumes[0].Name = deploymentsConfig.Spec.Template.Spec.Volumes[0].Name + "-" + fmt.Sprintf("%d", envNum)
	deploymentsConfig.Spec.Template.Spec.Volumes[0].ConfigMap.Name = deploymentsConfig.Spec.Template.Spec.Volumes[0].ConfigMap.Name + "-" + fmt.Sprintf("%d", envNum)
	deploymentsConfig.Spec.Template.Spec.Containers[0].VolumeMounts[0].Name = deploymentsConfig.Spec.Template.Spec.Containers[0].VolumeMounts[0].Name + "-" + fmt.Sprintf("%d", envNum)

	_, err = deploymentsClient.Create(context.TODO(), &deploymentsConfig, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func createOSDep(envNum int) error {

	deploymentsConfig := appv1.Deployment{}

	filename := resourcePath + "/" + osDepName

	deploymentsConfigFile, err := os.Open(filename)
	if err != nil {
		return err
	}

	decoder := json.NewDecoder(deploymentsConfigFile)

	err = decoder.Decode(&deploymentsConfig)
	if err != nil {
		return err
	}

	deploymentsConfig.Name = deploymentsConfig.Name + "-" + fmt.Sprintf("%d", envNum)
	deploymentsConfig.Spec.Selector.MatchLabels["app"] = deploymentsConfig.Spec.Selector.MatchLabels["app"] + "-" + fmt.Sprintf("%d", envNum)
	deploymentsConfig.Spec.Template.ObjectMeta.Labels["app"] = deploymentsConfig.Spec.Template.ObjectMeta.Labels["app"] + "-" + fmt.Sprintf("%d", envNum)
	_, err = deploymentsClient.Create(context.TODO(), &deploymentsConfig, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}

	return nil
}

func createConfigMap(envNum int) error {

	configMap, err := newNgxConfigMap(envNum)
	if err != nil {
		return err
	}

	_, err = configMapClient.Create(context.TODO(), configMap, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func getAllSvc() ([]v1.Service, error) {
	svcClients := clientSet.CoreV1().Services(v1.NamespaceDefault)
	svc, err := svcClients.List(context.TODO(), metav1.ListOptions{})
	envSvc := make([]v1.Service, 0)

	for _, item := range svc.Items {
		if strings.HasPrefix(item.Name, "os-svc") {
			envSvc = append(envSvc, item)
		}
	}

	return envSvc, err
}

func getAllEnv() ([]int, error) {
	services, err := getAllSvc()
	if err != nil {
		return nil, err
	}

	envNums := make([]int, 0)

	for _, item := range services {

		envNum, err := svc2envNum(item.Name)
		if err != nil {
			return envNums, err
		}

		envNums = append(envNums, envNum)
	}

	return envNums, nil
}

func listEnv() {
	envNums, err := getAllEnv()
	if err != nil {
		panic(err)
	}

	fmt.Print("envNum")

	for _, num := range envNums {
		fmt.Print("\n", num)
	}

	fmt.Println()
}

func svc2envNum(svcName string) (int, error) {
	svcEnvNum := strings.TrimPrefix(svcName, "os-svc-")
	envNum, err := strconv.Atoi(svcEnvNum)
	if err != nil {
		return 0, err
	}
	return envNum, nil
}

func initEnvMessage() {
	envNums, err := getAllEnv()
	if err != nil {
		panic(err)
	}

	for _, num := range envNums {
		currentMaxEnv = max(currentMaxEnv, num)
	}

}
