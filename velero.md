# [TBD] Back up Wordpress with Velero
For this part we will use [velero.io](https://velero.io) previously known as Heprio Ark
## Prerequisites:

1. Azure Subscription
2. Helm - [installation guide](https://helm.sh/docs/using_helm/#installing-helm)
3. [Azure CLI installed](https://docs.microsoft.com/en-us/cli/azure/install-azure-cli?view=azure-cli-latest)

## Create an AKS cluster

```bash
export AKS_RG=...
export AKS_NAME=...
export VERSION=$(az aks get-versions -l westeurope --query 'orchestrators[-1].orchestratorVersion' -o tsv)

az group create --name $AKS_RG --location westeurope
```

```bash
az aks create --resource-group $AKS_RG \
 --name $AKS_NAME \
 --location westeurope \
 --kubernetes-version $VERSION \
 --generate-ssh-keys \
 --disable-rbac

 az aks get-credentials --resource-group $AKS_RG --name $AKS_NAME
```

## Download Velero

1. Download the latest [official release's](https://github.com/heptio/velero/releases) tarball for your client platform.
2. Extract the tarball:
```bash
tar -xvf <RELEASE-TARBALL-NAME>.tar.gz -C /dir/to/extract/to
```
3. Move the `velero` binary from the Velero directory to somewhere in your PATH

## Initialize helm
Inside the cloned repo run:
```bash
kubectl apply -f helm-demo/helm-rbac.yaml
helm init --service-account tiller
kubectl get pods -n kube-system | grep tiller
```

## Create Azure storage account and blob container

```bash
export AZURE_BACKUP_RESOURCE_GROUP=velero-backup
az group create -n $AZURE_BACKUP_RESOURCE_GROUP --location westeurope

export AZURE_STORAGE_ACCOUNT_ID="velero$(uuidgen | cut -d '-' -f5 | tr '[A-Z]' '[a-z]')"
az storage account create \
    --name $AZURE_STORAGE_ACCOUNT_ID \
    --resource-group $AZURE_BACKUP_RESOURCE_GROUP \
    --sku Standard_GRS \
    --encryption-services blob \
    --https-only true \
    --kind BlobStorage \
    --access-tier Hot
```
Create the blob container named velero. Feel free to use a different name, preferably unique to a single Kubernetes cluster. See the FAQ for more details.

```bash
export BLOB_CONTAINER=velero
az storage container create -n $BLOB_CONTAINER --public-access off --account-name $AZURE_STORAGE_ACCOUNT_ID
```
## Get resource group for persistent volume shanpshots

AKS stores all the resources in a separate resource group which name starts with `MC_*`. Locate this group and store its name in a variable:
```bash
az group list
export AZURE_RESOURCE_GROUP=...
```

## Create service principal

1. Obtain your Azure Account Subscription ID and Tenant ID
```bash
export AZURE_SUBSCRIPTION_ID=`az account list --query '[?isDefault].id' -o tsv`
export AZURE_TENANT_ID=`az account list --query '[?isDefault].tenantId' -o tsv`
```
2. Create a service principal with Contributor role. This will have subscription-wide access, so protect this credential. Create service principal and let the CLI generate a password for you. Make sure to capture the password:

```bash
export AZURE_CLIENT_SECRET=`az ad sp create-for-rbac --name "velero$AZURE_STORAGE_ACCOUNT_ID" --role "Contributor" --query 'password' -o tsv`
export AZURE_CLIENT_ID=`az ad sp list --display-name "velero" --query '[0].appId' -o tsv`
```

3. Now you need to create a file that contains all the environment variables you just set. The command looks like the following:

```bash
cat << EOF  > ./credentials-velero
AZURE_SUBSCRIPTION_ID=${AZURE_SUBSCRIPTION_ID}
AZURE_TENANT_ID=${AZURE_TENANT_ID}
AZURE_CLIENT_ID=${AZURE_CLIENT_ID}
AZURE_CLIENT_SECRET=${AZURE_CLIENT_SECRET}
AZURE_RESOURCE_GROUP=${AZURE_RESOURCE_GROUP}
EOF
```
## Install and start Velero

```bash
velero install \
    --provider azure \
    --bucket $BLOB_CONTAINER \
    --secret-file ./credentials-velero \
    --backup-location-config resourceGroup=$AZURE_BACKUP_RESOURCE_GROUP,storageAccount=$AZURE_STORAGE_ACCOUNT_ID \
    --snapshot-location-config apiTimeout=30
```
## Edit the BackupStorageLocation in velero namespace
Change the `resourceGroup` parameter to `$AZURE_BACKUP_RESOURCE_GROUP` and `storageAccount` to `$AZURE_STORAGE_ACCOUNT_ID`
```bash
kubectl edit backupstoragelocation default -n velero
```

## Install Wordpress and get a password

```bash
helm install stable/wordpress --namespace wordpress --name wordpress-velero
kubectl get secret --namespace wordpress wordpress-velero -o jsonpath="{.data.wordpress-password}" | base64 --decode
```
## Perform a manual backup

```bash
velero backup create wordpress --include-namespaces wordpress --wait
```
TBD and correct get secret command