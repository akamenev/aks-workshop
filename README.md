# aks-workshop
Repo for AKS Workshop content

## Building the app image
### Running the app with go run
```
go run main.go
```

### Building the image with golang base image
```
docker pull golang
docker build -t welcome-app-golang:golang .
docker run -it -p 8080:8080 welcome-app-golang:golang
```

### Building the image with scratch base image
```
go build -o main.out .
docker build -t welcome-app-golang:scratch -f Dockerfile.scratch
docker run -it -p 8080:8080 welcome-app-golang:scratch
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main.out .
```