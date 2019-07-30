## Creating Pods, Deployments, Services

### Prerequisites:

1. Azure Subscription
2. Completed steps from [Docker Image Creation, Azure Container Registry, Azure Container Instances](https://github.com/akamenev/aks-workshop/blob/master/docker-images-acr-aci.md)

### Create a managed Kubernetes cluster (AKS)

```bash
export AKS_RG=...
export AKS_NAME=..

az group create --name $AKS_RG --location westeurope
az aks create --resource-group $AKS_RG --name $AKS_NAME --generate-ssh-keys
az aks install-cli
az aks get-credentials --name $AKS_NAME --resource-gorup $AKS_RG
kubectl get nodes
```

### Create a simple pod
Inside the cloned repo go to a `kube-concepts` folder, create a pod and verify that it is running:

```bash
cd kube-concepts
kubectl apply -f simple-pod-nginx.yaml
kubectl get pods
```

### Create a secret to store ACR credentials

```bash
export EMAIL=...

kubectl create secret docker-registry $REGISTRY_NAME --docker-server=$REGISTRY_NAME.azurecr.io --docker-username=$REGISTRY_NAME --docker-password=$PASSWORD --docker-email=$EMAIL
```

### Deploy a welcome-app and scale it
1. Open `welcome-app-deployment.yaml` and replace REGISTRY_NAME with your registry name:
```yaml
...
      containers:
      - name: welcome-app
        image: REGISTRY_NAME.azurecr.io/welcome-app:v1
        ports:
        - containerPort: 8080
        resources:
          requests:
            cpu: 100m
            memory: 128Mi
          limits:
            cpu: 150m
            memory: 128Mi
      imagePullSecrets:
      - name: REGISTRY_NAME
```
2. Create welcome-app deployment and service
```bash
kubectl apply -f welcome-app-deployment.yaml
kubectl apply -f welcome-app-service.yaml
```
3. Retrieve an external service IP and go to a wecome-app at `http://public_ip`. Note: it can take several minutes for AKS to obtain a public IP
```
kubectl get svc
```

4. Scale a welcome-app deployment
```bash
kubectl scale deployment welcome-app --replicas=4
kubectl get pods
```