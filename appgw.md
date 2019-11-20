# Creating Ingress

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