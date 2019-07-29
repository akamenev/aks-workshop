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