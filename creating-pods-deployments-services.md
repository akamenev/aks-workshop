## Creating Pods, Deployments, Services

### Prerequisites:

1. Azure Subscription
2. Completed steps from [Docker Image Creation, Azure Container Registry, Azure Container Instances](https://github.com/akamenev/aks-workshop/blob/master/docker-images-acr-aci.md)

### Create a managed Kubernetes cluster (AKS)
```bash
export AKS_RG=...
export AKS_NAME=...

az group create --name $AKS_RG --location westeurope
az aks create --resource-group $AKS_RG --name $AKS_NAME --generate-ssh-keys
az aks install-cli
az aks get-credentials --name $AKS_NAME --resource-gorup $AKS_RG
```

```bash
kubectl apply -f simple-pod-nginx.yaml
```
```bash
kubectl create secret docker-registry kamenevlabs --docker-server=kamenevlabs.azurecr.io --docker-username=kamenevlabs --docker-password=$PASSWORD --docker-email=EMAIL
```
```bash
kubectl apply -f welcome-app-deployment.yaml
kubectl apply -f welcome-apps-service.yaml
kubectl get svc
```
```bash
kubectl scale deployment welcome-app --replicas=4
```