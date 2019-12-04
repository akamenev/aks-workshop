# Deploying Application Gateway Ingress Controller

The Application Gateway Ingress Controller allows Azure Application Gateway to be used as the ingress for an Azure Kubernetes Service aka AKS cluster.

As shown in the figure below, the ingress controller runs as a pod within the AKS cluster. It consumes Kubernetes Ingress Resources and converts them to an Azure Application Gateway configuration which allows the gateway to load-balance traffic to Kubernetes pods

![image](appgw/appgw.png)

More info about Application Gateway Ingress Controller can be found [here](https://azure.github.io/application-gateway-kubernetes-ingress/).
## Prerequisites:

1. Azure Subscription
2. Helm 3 - [installation guide](https://helm.sh/docs/using_helm/#installing-helm)
3. [Azure CLI installed](https://docs.microsoft.com/en-us/cli/azure/install-azure-cli?view=azure-cli-latest)

## Create an AKS cluster with Azure CNI

```bash
export AKS_RG=...
export AKS_NAME=...
export VERSION=$(az aks get-versions -l westeurope --query 'orchestrators[-1].orchestratorVersion' -o tsv)

az group create --name $AKS_RG --location westeurope

az network vnet create --resource-group $AKS_RG --name $AKS_NAME --address-prefixes 10.0.0.0/8 --subnet-name akssubnet --subnet-prefix 10.240.0.0/16
```
Create a service principal and read in the application ID and password
```bash
SP_PASSWORD=$(az ad sp create-for-rbac --name aksagic$AKS_NAME --query 'password' -o tsv)
SP_ID=$(az ad sp list --display-name "aksagic$AKS_NAME" --query '[0].appId' -o tsv)

```
```bash
SUBNET_ID=$(az network vnet subnet show --resource-group $AKS_RG --vnet-name $AKS_NAME --name akssubnet --query id -o tsv)
```
Create cluster
```bash
az aks create \
    --resource-group $AKS_RG \
    --name $AKS_NAME \
    --node-count 3 \
    --generate-ssh-keys \
    --network-plugin azure \
    --service-cidr 10.0.0.0/16 \
    --dns-service-ip 10.0.0.10 \
    --docker-bridge-address 172.17.0.1/16 \
    --vnet-subnet-id $SUBNET_ID \
    --service-principal $SP_ID \
    --client-secret $SP_PASSWORD

az aks get-credentials --resource-group $AKS_RG --name $AKS_NAME
```

## Create Application Gateway
Create a subnet and public IP address for the Application Gateway
```bash
az network vnet subnet create \
  --name appgw \
  --resource-group $AKS_RG \
  --vnet-name $AKS_NAME   \
  --address-prefix 10.0.2.0/24

az network public-ip create \
  --resource-group $AKS_RG \
  --name appgwPublicIP \
  --allocation-method Static \
  --sku Standard
```
Create an Application Gateway
```bash
az network application-gateway create \
  --name $AKS_NAME \
  --location westeurope \
  --resource-group $AKS_RG \
  --capacity 2 \
  --sku WAF_v2 \
  --public-ip-address appgwPublicIP \
  --vnet-name $AKS_NAME \
  --subnet appgw
```
Create an Azure identity
```bash
az identity create -g $AKS_RG -n $AKS_NAME
```
For the role assignment commands below we need to obtain `principalId` for the newly created identity:
```bash
export IDENTITY_PID=$(az identity show -g $AKS_RG -n $AKS_NAME --query 'principalId' -o tsv)
```
Give the identity `Contributor` access to your App Gateway. For this you need the ID of the App Gateway:
```bash
export APPGW_ID=$(az network application-gateway list --resource-group $AKS_RG --query '[].id' -o tsv)

az role assignment create \
    --role Contributor \
    --assignee $IDENTITY_PID \
    --scope $APPGW_ID
```
Give the identity `Reader` access to the App Gateway resource group:
```bash
export RG_ID=$(az group list --query '[].id' -o tsv | grep $AKS_RG | head -n 1)

az role assignment create \
    --role Reader \
    --assignee $IDENTITY_PID \
    --scope $RG_ID
```
## Install Ingress Controller as a Helm Chart

Add the `application-gateway-kubernetes-ingress' helm repo and perform helm update:
```bash
helm repo add application-gateway-kubernetes-ingress https://appgwingress.blob.core.windows.net/ingress-azure-helm-package/

helm repo update
```

In the `appgw` directory open and edit `helm-config.yaml` or use `sed` commands below:
```bash
export SUBSCRIPTION_ID=...
export IDENTITY_RID=$(az identity show -g $AKS_RG -n $AKS_NAME --query 'id' -o tsv)
export IDENTITY_CID=$(az identity show -g $AKS_RG -n $AKS_NAME --query 'clientId' -o tsv)

sed -i "s/SUBSCRIPTION_ID/$SUBSCRIPTION_ID/g" helm-config.yaml
sed -i "s/AKS_NAME/$AKS_NAME/g" helm-config.yaml
sed -i "s/AKS_RG/$AKS_RG/g" helm-config.yaml
sed -i "s%IDENTITY_CID%$IDENTITY_CID%g" helm-config.yaml
sed -i "s%IDENTITY_RID%$IDENTITY_RID%g" helm-config.yaml

```
Install the Helm chart
```bash
export AGIC_NAMESPACE=default

kubectl apply -f pod-identity-rbac-deployment.yaml

helm install agic-ingress -f helm-config.yaml \
application-gateway-kubernetes-ingress/ingress-azure \
--set appgw.usePrivateIP=false
```

## Deploy the Voting and Welcome apps
```bash
kubectl apply -f voting-app-all.yaml
kubectl apply -f welcome-app-deployment.yaml
kubectl apply -f welcome-app-service.yaml
```

## Create ingress for the apps
Get the public IP address of the Application Gateway
```bash
export APPGW_IP=$(az network public-ip show --resource-group $AKS_RG --name appgwPublicIP --query 'ipAddress' -o tsv)
```
Replace APPGWIP in `voting-app-ingress.yaml` and `welcome-app-ingress.yaml` with your value or by using `sed` command below
```bash
sed -i "s/APPGW_IP/$APPGW_IP/g" welcome-app-ingress.yaml
sed -i "s/APPGW_IP/$APPGW_IP/g" voting-app-ingress.yaml
```
Apply ingress manifests
```bash
kubectl apply -f welcome-app-ingress.yaml
kubectl apply -f voting-app-ingress.yaml
```

Check `vote.APPGW_IP.nip.io` and `result.APPGW_IP.nip.io` to see the Voting app and `welcome.APPGW_IP.nip.io` to see the Welcome app