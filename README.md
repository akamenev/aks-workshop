# AKS Workshop
Repo for AKS Workshop content

## Content:


1. [Docker Image Creation, Azure Container Registry, Azure Container Instances](https://github.com/akamenev/aks-workshop-01-Docker-Images-ACR-ACI.md)

## Creating Pods, Deployments, Services
```
kubectl apply -f simple-pod-nginx.yaml
```
```
kubectl create secret docker-registry kamenevlabs --docker-server=kamenevlabs.azurecr.io --docker-username=kamenevlabs --docker-password=$password --docker-email=EMAIL
```
```
kubectl apply -f welcome-app-deployment.yaml
kubectl apply -f welcome-apps-service.yaml
kubectl get svc
```
```
kubectl scale deployment welcome-app --replicas=4
```

## Helm, creating, deploying and pushing a chart
```
helm init
kubectl get pods -n kube-system | grep tiller
```
```
helm create welcome-app
helm install --name welcome-app --namespace welcome-app ./welcome-app
kubectl get all -n welcome-app
helm ls
helm delete welcome-app
```
```
helm package
az acr helm push welcome-app-0.1.0.tgz --name kamenevlabs
az acr helm repo add --name kamenevlabs
helm search welcome-app
az acr helm list --name kamenevlabs
```
```
helm repo update
helm install kamenevlabs/welcome-app
```
```
helm fetch stable/wordpress --untar
```