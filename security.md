# [TBD] Security

## Prerequisites:

1. Azure Subscription
2. [Azure CLI installed](https://docs.microsoft.com/en-us/cli/azure/install-azure-cli?view=azure-cli-latest)

## Register pod security policy feature provider

```bash
az feature register --name PodSecurityPolicyPreview --namespace Microsoft.ContainerService
```
It takes a few minutes for the status to show Registered. You can check on the registration status using the `az feature list` command:
```bash
az feature list -o table --query "[?contains(name, 'Microsoft.ContainerService/PodSecurityPolicyPreview')].{Name:name,State:properties.state}"
```
When ready, refresh the registration of the Microsoft.ContainerService resource provider using the az provider register command:
```bash
az provider register --namespace Microsoft.ContainerService
```

## Create an AKS cluster

```bash
export AKS_RG=...
export AKS_NAME=...
export VERSION=$(az aks get-versions -l westeurope --query 'orchestrators[-1].orchestratorVersion' -o tsv)

az group create --name $AKS_RG --location westeurope

az network vnet create --resource-group $AKS_RG --name $AKS_NAME --address-prefixes 10.0.0.0/8 --subnet-name akssubnet --subnet-prefix 10.240.0.0/16
```
Create a service principal and read in the application ID and password
```bash
SP_PASSWORD=$(az ad sp create-for-rbac --name akssec$AKS_NAME --query 'password' -o tsv)
SP_ID=$(az ad sp list --display-name "akssec$AKS_NAME" --query '[0].appId' -o tsv)

```
```bash
SUBNET_ID=$(az network vnet subnet show --resource-group $AKS_RG --vnet-name $AKS_NAME --name akssubnet --query id -o tsv)
```
For real-world use, don't enable the pod security policy until you have defined your own custom policies. In this article, you enable pod security policy as the first step to see how the default policies limit pod deployments.
```bash
az aks create \
    --resource-group $AKS_RG \
    --name $AKS_NAME \
    --node-count 1 \
    --generate-ssh-keys \
    --network-plugin azure \
    --service-cidr 10.0.0.0/16 \
    --dns-service-ip 10.0.0.10 \
    --docker-bridge-address 172.17.0.1/16 \
    --vnet-subnet-id $SUBNET_ID \
    --service-principal $SP_ID \
    --client-secret $SP_PASSWORD \
    --network-policy azure \
    --enable-pod-security-policy

az aks get-credentials --resource-group $AKS_RG --name $AKS_NAME
```

## Secure pod traffic with network policies

### Deny all inbound traffic to a pod
Before you define rules to allow specific network traffic, first create a network policy to deny all traffic. This policy gives you a starting point to begin to whitelist only the desired traffic. You can also clearly see that traffic is dropped when the network policy is applied.

For the sample application environment and traffic rules, let's first create a namespace called development to run the example pods:
```bash
kubectl create namespace development
kubectl label namespace/development purpose=development
```

Create an example back-end pod that runs NGINX. This back-end pod can be used to simulate a sample back-end web-based application. Create this pod in the development namespace, and open port 80 to serve web traffic. Label the pod with app=webapp,role=backend so that we can target it with a network policy in the next section:

```bash
kubectl run backend --image=nginx --labels app=webapp,role=backend --namespace development --expose --port 80 --generator=run-pod/v1
```
Create another pod and attach a terminal session to test that you can successfully reach the default NGINX webpage:
```bash
kubectl run --rm -it --image=alpine network-policy --namespace development --generator=run-pod/v1
```
At the shell prompt, use wget to confirm that you can access the default NGINX webpage:
```bash
wget -qO- http://backend
exit
```

#### Create and apply a network policy
Now that you've confirmed you can use the basic NGINX webpage on the sample back-end pod, create a network policy to deny all traffic.

Inside the cloned repo go to a `security` folder and apply the `backend-policy.yaml`:
```bash
kubectl apply -f backend-policy.yaml
```

#### Test the network policy
Let's see if you can use the NGINX webpage on the back-end pod again. Create another test pod and attach a terminal session:
```bash
kubectl run --rm -it --image=alpine network-policy --namespace development --generator=run-pod/v1
```
At the shell prompt, use wget to see if you can access the default NGINX webpage. This time, set a timeout value to 2 seconds. The network policy now blocks all inbound traffic, so the page can't be loaded, as shown in the following example:
```bash
wget -qO- --timeout=2 http://backend
exit
```

### Allow inbound traffic based on pod label
In the previous section, a back-end NGINX pod was scheduled, and a network policy was created to deny all traffic. Let's create a front-end pod and update the network policy to allow traffic from front-end pods.

Update the network policy to allow traffic from pods with the labels app:webapp,role:frontend and in any namespace:
```bash
kubectl apply -f backend-policy-label.yaml
```
Schedule a pod that is labeled as app=webapp,role=frontend and attach a terminal session:
```bash
kubectl run --rm -it frontend --image=alpine --labels app=webapp,role=frontend --namespace development --generator=run-pod/v1
```
At the shell prompt, use wget to see if you can access the default NGINX webpage:
```bash
wget -qO- http://backend
exit
```
Because the ingress rule allows traffic with pods that have the labels app: webapp,role: frontend, the traffic from the front-end pod is allowed.

#### Test a pod without a matching label
The network policy allows traffic from pods labeled app: webapp,role: frontend, but should deny all other traffic. Let's test to see whether another pod without those labels can access the back-end NGINX pod. Create another test pod and attach a terminal session:
```bash
kubectl run --rm -it --image=alpine network-policy --namespace development --generator=run-pod/v1
```
At the shell prompt, use wget to see if you can access the default NGINX webpage:
```bash
wget -qO- --timeout=2 http://backend
exit
```

### Allow traffic only from within a defined namespace
In the previous examples, you created a network policy that denied all traffic, and then updated the policy to allow traffic from pods with a specific label. Another common need is to limit traffic to only within a given namespace. If the previous examples were for traffic in a development namespace, create a network policy that prevents traffic from another namespace, such as production, from reaching the pods.

First, create a new namespace to simulate a production namespace:
```bash
kubectl create namespace production
kubectl label namespace/production purpose=production
```
Schedule a test pod in the production namespace that is labeled as app=webapp,role=frontend. Attach a terminal session:
```bash
kubectl run --rm -it frontend --image=alpine --labels app=webapp,role=frontend --namespace production --generator=run-pod/v1
```
At the shell prompt, use wget to confirm that you can access the default NGINX webpage:
```bash
wget -qO- http://backend.development
exit
```
Because the labels for the pod match what is currently permitted in the network policy, the traffic is allowed. The network policy doesn't look at the namespaces, only the pod labels.

#### Update the network policy
Let's update the ingress rule namespaceSelector section to only allow traffic from within the development namespace:
```bash
kubectl apply -f backend-policy-namespaces.yaml
```

#### Test the updated network policy
Schedule another pod in the production namespace and attach a terminal session:
```bash
kubectl run --rm -it frontend --image=alpine --labels app=webapp,role=frontend --namespace production --generator=run-pod/v1
```
At the shell prompt, use wget to see that the network policy now denies traffic:
```bash
wget -qO- --timeout=2 http://backend.development
exit
```

### Clean up resources
```bash
kubectl delete namespace production
kubectl delete namespace development
```