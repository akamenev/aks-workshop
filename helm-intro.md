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