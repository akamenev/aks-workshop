## Autoscaling in AKS
This part is based on aksworkshop.io
### Prerequisites:

1. Azure Subscription
2. Helm - [installation guide](https://helm.sh/docs/using_helm/#installing-helm)

### Create an AKS cluster with the cluster autoscaler

```bash
export AKS_RG=...
export AKS_NAME=...
export VERSION=$(az aks get-versions -l westeurope --query 'orchestrators[-1].orchestratorVersion' -o tsv)

az group create --name $AKS_RG --location westeurope
```

AKS clusters that support the cluster autoscaler must use virtual machine scale sets and run Kubernetes version 1.12.4 or later. This scale set support is in preview. 
<details><summary> Opt in preview features </summary>
<p>
To opt in and create clusters that use scale sets, first install the aks-preview Azure CLI extension using the `az extension add` command, as shown in the following example:

```bash
az extension add --name aks-preview
```
To create an AKS cluster that uses scale sets, you must also enable a feature flag on your subscription. To register the VMSSPreview feature flag, use the `az feature register` command as shown in the following example:
```bash
az feature register --name VMSSPreview --namespace Microsoft.ContainerService
```
It takes a few minutes for the status to show Registered. You can check on the registration status using the `az feature list` command:
```bash
az feature list -o table --query "[?contains(name, 'Microsoft.ContainerService/VMSSPreview')].{Name:name,State:properties.state}"
```
When ready, refresh the registration of the Microsoft.ContainerService resource provider using the `az provider register` command:
```bash
az provider register --namespace Microsoft.ContainerService
```
</p>
</details>

Use the az aks create command specifying the --enable-cluster-autoscaler parameter, and a node --min-count and --max-count
```bash
az aks create --resource-group $AKS_RG \
 --name $AKS_NAME \
 --location westeurope \
 --kubernetes-version $VERSION \
 --generate-ssh-keys \
 --enable-vmss \
 --enable-cluster-autoscaler \
 --min-count 1 \
 --max-count 3

 az aks get-credentials --resource-group $AKS_RG --name $AKS_NAME
```

### Initialize helm
Inside the cloned repo go to `scaling` folder and run:
```bash
kubectl apply -f helm-rbac.yaml
helm init --service-account tiller
kubectl get pods -n kube-system | grep tiller
```

### Deploy MongoDB
After you have Tiller initialized in the cluster, wait for a short while then install the MongoDB chart, then take note of the username, password and endpoints created. The command below creates a user called orders-user and a password of orders-password
```bash
helm install stable/mongodb --name orders-mongo --set mongodbUsername=orders-user,mongodbPassword=orders-password,mongodbDatabase=akschallenge
```
In the previous step, you installed MongoDB using Helm, with a specified username, password and a hostname where the database is accessible. You’ll now create a Kubernetes secret called `mongodb` to hold those details, so that you don’t need to hard-code them in the YAML files.
```bash
kubectl create secret generic mongodb --from-literal=mongoHost="orders-mongo-mongodb.default.svc.cluster.local" --from-literal=mongoUser="orders-user" --from-literal=mongoPassword="orders-password"
```

### Deploy Order Capture API
You need to deploy the Order Capture API `(azch/captureorder)`. This requires an external endpoint, exposing the API on port 80 and needs to write to MongoDB.
```bash
kubectl apply -f captureorder-deployment.yaml
kubectl apply -f captureorder-service.yaml
```
Retrieve the External-IP of the Service. Use the command below. Make sure to allow a couple of minutes for the Azure Load Balancer to assign a public IP.
```bash
kubectl get service captureorder -o jsonpath="{.status.loadBalancer.ingress[*].ip}"
```
Ensure order are succesfully written to MongoDB. Send a POST request using Postman or curl to the IP of the service you got from the previous command. You can expect the order ID returned by API once your order has been written into Mongo DB successfully
```bash
export API_IP=$(kubectl get service captureorder -o jsonpath="{.status.loadBalancer.ingress[*].ip}")

curl -d '{"EmailAddress": "email@domain.com", "Product": "prod-1", "Total": 100}' -H "Content-Type: application/json" -X POST http://$API_IP/v1/order
```

### Run a baseline load test

There is a a container image on Docker Hub `(azch/loadtest)` that is preconfigured to run the load test. You may run it in Azure Container Instances running the command below
```bash
export RG_LOAD=...

az group create --name $RG_LOAD -l westeurope
az container create -g $RG_LOAD -n loadtest --image azch/loadtest --restart-policy Never -e SERVICE_IP=$API_IP
```
This will fire off a series of increasing loads of concurrent users (100, 400, 1600, 3200, 6400) POSTing requests to your Order Capture API endpoint with some wait time in between to simulate an increased pressure on your application.

You may view the logs of the Azure Container Instance streaming logs by running the command below. You may need to wait for a few minutes to get the full logs, or run this command multiple times.
```bash
az container logs -g $RG_LOAD -n loadtest
```
When you’re done, you may delete it by running
```bash
az container delete -g $RG_LOAD -n loadtest
```
Make note of results (sample below), figure out what is the breaking point for the number of users.

### Create Horizontal Pod Autoscaler

Horizontal Pod Autoscaler allows Kubernetes to detect when your deployed pods need more resources and then it schedules more pods onto the cluster to cope with the demand.
See the `captureorder-hpa.yaml` in the `scaling` folder and deploy it:
```bash
kubectl apply -f captureorder-hpa.yaml
```
Run the load test again
```bash
az container create -g $RG_LOAD -n loadtest --image azch/loadtest --restart-policy Never -e SERVICE_IP=$API_IP
```
Observe your Kubernetes cluster reacting to the load by running
```bash
watch kubectl get pods -l  app=captureorder
```
Since you configured your AKS cluster with cluster autoscaler, you should see it dynamically adding and removing nodes based on the cluster utilization. To change the node count, use the az aks update command and specify a minimum and maximum value. The following example sets the `--min-count` to 1 and the `--max-count` to 5:
```bash
az aks update \
  --resource-group $AKS_RG \
  --name $AKS_NAME \
  --update-cluster-autoscaler \
  --min-count 1 \
  --max-count 5
```