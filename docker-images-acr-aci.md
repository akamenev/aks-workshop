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