

all:
	kubectl apply -f deployment.yaml
	kubectl apply -f ingress.yaml

clean:
	kubectl delete -f ingress.yaml
	kubectl delete -f deployment.yaml