# AKS Workshop
Repo for AKS Workshop content

## Docker Image Creation, Azure Container Registry, Azure Container Instances
### Running the app with go run
```
go run main.go
```

### Building the image with golang base image
```
docker pull golang
docker build -t welcome-app-golang:v1 .
docker run -it -p 8080:8080 welcome-app-golang:v1
```

### Building the image with scratch base image
```
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main.out .
docker build -t welcome-app-scratch:v1 -f Dockerfile.scratch .
docker run -it -p 8080:8080 welcome-app-scratch:v1
```
### Push the image to the registry
```
docker tag welcome-app-scratch:v1 kamenevlabs.azurecr.io/welcome-app:v1
az login
az acr login --name kamenevlabs
docker push kamenevlabs.azurecr.io/welcome-app:v1
```

### Build in Azure
```
az acr build --registry kamenevlabs --image welcome-app-acrbuild:v1 -f Dockerfile.scratch .
```

### Run in container instance
```
password=$(az acr credential show-n kamenevlabs --query "passwords[0].value" -o tsv)
az group create -n welcome-app -l westeurope
az container create -n welcome-app -g welcome-app --image "kamenevlabs.azurecr.io/welcome-app:v1" --registry-username "kamenevlabs" --registry-password "$password" --ports 8080 --dns-name-label welcomeapptest
```

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