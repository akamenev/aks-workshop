## Docker Image Creation, Azure Container Registry, Azure Container Instances

### Prerequisites:

1. Azure Subscription
2. [Golang](https://golang.org/doc/install) installed
3. Local Docker installed ([Linux](https://docs.docker.com/install/linux/docker-ce/ubuntu/#install-docker-ce), [Mac](https://docs.docker.com/docker-for-mac/install/) or [Windows](https://docs.docker.com/docker-for-windows/install/))

### Clone the repo
```
git clone https://github.com/akamenev/aks-workshop.git
```

### Running the app with go run
Inside the cloned repo go to a welcome-app folder and run the app
```
cd welcome-app
go run main.go
```
Open `http://localhost:8080` in your browser and see the welcome app screen. You can add your name as a GET parameter to change the message - `http://localhost:8080?name=Andrei`

### Building the image with golang base image
Pull the golang image from Dockerhub, build the container and run it
```
docker pull golang
docker build -t welcome-app-golang:v1 .
docker run -it -p 8080:8080 welcome-app-golang:v1
```
Try to open a welcome-app again at `http://localhost:8080`

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