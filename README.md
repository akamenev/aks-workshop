# aks-workshop
Repo for AKS Workshop content

## Building the app image

```
go run main.go
```
```
docker pull golang
docker build -t welcome-app-golang:golang .
docker run -it -p 8080:8080 welcome-app-golang:golang
```
```
go build -o main.out .
docker build -t welcome-app-golang:scratch -f Dockerfile.scratch
docker run -it -p 8080:8080 welcome-app-golang:scratch
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main.out .
```