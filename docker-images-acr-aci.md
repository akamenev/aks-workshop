## Docker Image Creation, Azure Container Registry, Azure Container Instances

### Prerequisites:

1. Azure Subscription
2. [Golang](https://golang.org/doc/install) installed
3. Local Docker installed ([Linux](https://docs.docker.com/install/linux/docker-ce/ubuntu/#install-docker-ce), [Mac](https://docs.docker.com/docker-for-mac/install/) or [Windows](https://docs.docker.com/docker-for-windows/install/))

### Clone the repo
```bash
git clone https://github.com/akamenev/aks-workshop.git
```

### Running the app with go run
Inside the cloned repo go to a `welcome-app` folder and run the app
```bash
cd welcome-app
go run main.go
```
Open `http://localhost:8080` in your browser and see the welcome app screen. You can add your name as a GET parameter to change the message - `http://localhost:8080?name=Andrei`

### Building the image with golang base image
Pull the golang image from Dockerhub, build the container and run it
```bash
docker pull golang
docker build -t welcome-app-golang:v1 .
docker run -it -p 8080:8080 welcome-app-golang:v1
```
Try to open a welcome-app again at `http://localhost:8080`

### Building the image with scratch base image
Run the `docker images` command to see the welcome-app image size. It is more than 800 MB which is pretty big for a small app like that. For a compiled languages we can build a static binary that we can use with a `sratch` docker image as a base. Open the `Dockerfile` and `Dockerfile.scratch` files in editor to see the difference.

Build a Go binary and build a new docker image.

```bash
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main.out .

docker build -t welcome-app-scratch:v1 -f Dockerfile.scratch .
docker run -it -p 8080:8080 welcome-app-scratch:v1
```
Open a welcome-app again at `http://localhost:8080` and check the size of the image via `docker images`

### Create an Azure Container Registry
Specify Resource Group and Registry names and create ACR
```bash
export RG_NAME=...
export REGISTRY_NAME=...

az login
az group create -n $RG_NAME -l westeurope
az acr create --name $REGISTRY_NAME --resource-group $RG_NAME --sku basic --location westeurope --admin-enabled true
```

### Push the image to the registry
```bash
docker tag welcome-app-scratch:v1 $REGISTRY_NAME.azurecr.io/welcome-app:v1
az acr login --name $REGISTRY_NAME
docker push $REGISTRY_NAME.azurecr.io/welcome-app:v1
```

### Build in Azure
With ACR you don't even need a docker installed locally, you can build your images using ACR.
```bash
az acr build --registry $REGISTRY_NAME --image welcome-app-acrbuild:v1 -f Dockerfile.scratch .
```

### Run in container instance
Run welcome-app on Azure Container Instance
```bash
export PASSWORD=$(az acr credential show -n $REGISTRY_NAME --query "passwords[0].value" -o tsv)
export DNS_LABEL=...

az group create -n welcome-app -l westeurope
az container create -n welcome-app -g welcome-app --image "$REGISTRY_NAME.azurecr.io/welcome-app:v1" --registry-username "$REGISTRY_NAME" --registry-password "$PASSWORD" --ports 8080 --dns-name-label $DNS_LABEL
```
