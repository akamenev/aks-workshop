## Helm, creating, deploying and pushing a chart

### Prerequisites:

1. Azure Subscription
2. Completed steps from [Docker Image Creation, Azure Container Registry, Azure Container Instances](https://github.com/akamenev/aks-workshop/blob/master/docker-images-acr-aci.md)
3. Helm - [installation guide](https://helm.sh/docs/using_helm/#installing-helm)

### Initialize helm
Inside the cloned repo go to `helm-demo` folder and run:
```
kubectl apply -f helm-rbac.yaml
helm init --service-account tiller
kubectl get pods -n kube-system | grep tiller
```
### Edit the values.yaml file
Open `value.yaml` and `deployment.yaml` files inside the `welcome-app` and `welcome-app/templates` folder and change `REGISTRY_NAME` to your registry name or use the `sed` command below:
```yaml
...
image:
  repository: REGISTRY_NAME.azurecr.io/welcome-app-golang
  tag: v1
  pullPolicy: IfNotPresent
...
```
```yaml
...
 imagePullSecrets:
      - name: REGISTRY_NAME
...
```
```bash
sed -i "s/REGISTRY_NAME/$REGISTRY_NAME/g" welcome-app/values.yaml
sed -i "s/REGISTRY_NAME/$REGISTRY_NAME/g" welcome-app/templates/deployment.yaml
```
### Install the welcome-app via `helm install`
```bash
helm install --name welcome-app --namespace welcome-app ./welcome-app
helm ls
kubectl get all -n welcome-app

helm delete welcome-app
```

### Package the app and push it to the ACR Helm repository
```bash
helm package welcome-app

az acr login --name $REGISTRY_NAME
az acr helm push welcome-app-0.1.0.tgz --name $REGISTRY_NAME
```
### Deploy the app from the ACR Helm repository
```bash
az acr helm repo add --name $REGISTRY_NAME
helm search welcome-app
az acr helm list --name $REGISTRY_NAME
```
```bash
helm repo update
helm install $REGISTRY_NAME/welcome-app
kubect get pods
kubect get svc
```
### Explore the Wordpress Helm chart
Download and open the `wordpress` folder to explore the contents of a Wordpress chart
```bash
helm fetch stable/wordpress --untar
code wordpress
```