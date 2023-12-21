package main

import (
	"context"
	"encoding/json"
	"fmt"
	v1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"strconv"
)

const resourcePath = "/home/jking/GolandProjects/os-online/demo/resource"
const cmName = "cm.json"
const osDepName = "os-dep.json"
const ngxDepName = "ngx-dep.json"
const osSvcName = "os-svc.json"

const ingressName = "os-ingress"

var currentMaxEnv = 0

var ngxConf = string(`upstream codeserver {
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

}`)

// ingress struct should not be created by Ingress.Get()
// if use the struct, you will fail to create a new ingress
func createIngress() error {

	var tempIngress *netv1.Ingress

	oldIngress, err := clientSet.NetworkingV1().Ingresses(apiv1.NamespaceDefault).Get(context.Background(), ingressName, metav1.GetOptions{})
	if err != nil {
		tempIngress = osIngressT.DeepCopy()
	} else {
		tempIngress = new(netv1.Ingress)
		tempIngress.Spec = oldIngress.Spec
		tempIngress.Name = ingressName
	}

	pathType := new(netv1.PathType)
	*pathType = netv1.PathTypePrefix

	tempIngress.Spec.Rules[0].HTTP.Paths = append(tempIngress.Spec.Rules[0].HTTP.Paths, netv1.HTTPIngressPath{
		Path:     "/osenv/" + fmt.Sprintf("%d", currentMaxEnv),
		PathType: pathType,
		Backend: netv1.IngressBackend{
			Service: &netv1.IngressServiceBackend{
				Name: "os-svc-" + fmt.Sprintf("%d", currentMaxEnv),
				Port: netv1.ServiceBackendPort{
					Number: 80,
				},
			},
		},
	})

	// 先删除原来的ingress
	// 不存在就不删除
	if err != nil {
		_, err = clientSet.NetworkingV1().Ingresses(apiv1.NamespaceDefault).Create(context.Background(), tempIngress, metav1.CreateOptions{})
		if err != nil {
			return err
		}
	} else {
		err = clientSet.NetworkingV1().Ingresses(apiv1.NamespaceDefault).Delete(context.Background(), ingressName, metav1.DeleteOptions{})
		if err != nil {
			return err
		}
		_, err = clientSet.NetworkingV1().Ingresses(apiv1.NamespaceDefault).Create(context.Background(), tempIngress, metav1.CreateOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}

func deleteIngress(envNum int) error {

	oldIngress, err := clientSet.NetworkingV1().Ingresses(apiv1.NamespaceDefault).Get(context.Background(), ingressName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	tempIngress := osIngressT.DeepCopy()

	for i, v := range oldIngress.Spec.Rules[0].HTTP.Paths {
		if v.Path == "/osenv/"+fmt.Sprintf("%d", envNum) {
			if len(oldIngress.Spec.Rules[0].HTTP.Paths) == 1 {
				oldIngress.Spec.Rules[0].HTTP.Paths = make([]netv1.HTTPIngressPath, 0)
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
		err = clientSet.NetworkingV1().Ingresses(apiv1.NamespaceDefault).Delete(context.Background(), ingressName, metav1.DeleteOptions{})
		if err != nil {
			return err
		}
	} else {
		err = clientSet.NetworkingV1().Ingresses(apiv1.NamespaceDefault).Delete(context.Background(), ingressName, metav1.DeleteOptions{})
		if err != nil {
			return err
		}
		_, err = clientSet.NetworkingV1().Ingresses(apiv1.NamespaceDefault).Create(context.Background(), tempIngress, metav1.CreateOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}

func createOSSvc(envNum int) error {

	svcClients := clientSet.CoreV1().Services(apiv1.NamespaceDefault)

	filename := resourcePath + "/" + osSvcName

	svcConfig := apiv1.Service{}
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

	_, err = svcClients.Create(context.TODO(), &svcConfig, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func createNgxDep(envNum int) error {

	deploymentsClient := clientSet.AppsV1().Deployments(apiv1.NamespaceDefault)
	deploymentsConfig := v1.Deployment{}

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
	deploymentsClient := clientSet.AppsV1().Deployments(apiv1.NamespaceDefault)

	deploymentsConfig := v1.Deployment{}

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
	configMapClient := clientSet.CoreV1().ConfigMaps(apiv1.NamespaceDefault)

	configMap := apiv1.ConfigMap{}

	filename := resourcePath + "/" + cmName

	cmFile, err := os.Open(filename)
	if err != nil {
		return err
	}

	decoder := json.NewDecoder(cmFile)

	err = decoder.Decode(&configMap)
	if err != nil {
		return err
	}

	configMap.Name = configMap.Name + "-" + fmt.Sprintf("%d", envNum)
	configMap.Data["default.conf"] = fmt.Sprintf(ngxConf, "os-svc-"+fmt.Sprintf("%d", envNum), envNum)

	_, err = configMapClient.Create(context.TODO(), &configMap, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func listEnv() {
	svcClients := clientSet.CoreV1().Services(apiv1.NamespaceDefault)
	svc, err := svcClients.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}

	fmt.Println("envNum")

	for _, v := range svc.Items {
		if v.Name[:6] == "os-svc" {
			envNum, err := strconv.Atoi(v.Name[7:])
			if err != nil {
				panic(err)
			}
			fmt.Println("\n", envNum)
		}
	}
}

func initEnvMessage() {
	svcClients := clientSet.CoreV1().Services(apiv1.NamespaceDefault)
	svc, err := svcClients.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}

	for _, v := range svc.Items {
		if v.Name[:6] == "os-svc" {
			envNum, err := strconv.Atoi(v.Name[7:])
			if err != nil {
				panic(err)
			}
			currentMaxEnv = max(currentMaxEnv, envNum)
		}
	}

}
