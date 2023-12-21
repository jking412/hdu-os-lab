package main

import (
	apiv1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
)

// {
// "metadata": {
// "labels": {
// "app": "os-dep"
// }
// },
// "spec": {
// "containers": [
// {
// "name": "os",
// "image": "os:v3",
// "imagePullPolicy": "IfNotPresent",
// "ports": [
// {
// "containerPort": 8080,
// "protocol": "TCP",
// "name": "codeserver"
// }
// ],
// "env": [
// {
// "name": "CODE_SERVER_ARGS",
// "value": "--disable-workspace-trust --auth none"
// }
// ]
// }
// ]
// }
// }
var osPodT = &apiv1.Pod{}

var ngxPodT = &apiv1.Pod{}

var osSvcT = &apiv1.Service{}

var cmT = &apiv1.ConfigMap{}

// {
// "metadata": {
// "name": "os-ingress"
// },
// "spec": {
// "rules": [{
// "http": {
// "paths": []
// }
// }]
// }
// }
var osIngressT = &netv1.Ingress{
	Spec: netv1.IngressSpec{
		Rules: []netv1.IngressRule{
			{
				IngressRuleValue: netv1.IngressRuleValue{
					HTTP: &netv1.HTTPIngressRuleValue{
						Paths: []netv1.HTTPIngressPath{},
					},
				},
			},
		},
	},
}

func init() {
	osIngressT.Name = ingressName
}
