# Securing your AKS environment

This part is covering security topics in AKS. All the examples are from Microsoft docs:
* [Secure pod traffic with network policies](https://docs.microsoft.com/en-us/azure/aks/use-network-policies)
* [Use pod security policies](https://docs.microsoft.com/en-us/azure/aks/use-pod-security-policies)
For more accurate and up-to-date information please refer to these links.

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

## Use pod security policies
To improve the security of your AKS cluster, you can limit what pods can be scheduled. Pods that request resources you don't allow can't run in the AKS cluster. You define this access using pod security policies. This article shows you how to use pod security policies to limit the deployment of pods in AKS.

PodSecurityPolicy is an admission controller that validates a pod specification meets your defined requirements. These requirements may limit the use of privileged containers, access to certain types of storage, or the user or group the container can run as. When you try to deploy a resource where the pod specifications don't meet the requirements outlined in the pod security policy, the request is denied. This ability to control what pods can be scheduled in the AKS cluster prevents some possible security vulnerabilities or privilege escalations.

When you enable pod security policy in an AKS cluster, some default policies are applied. These default policies provide an out-of-the-box experience to define what pods can be scheduled. However, cluster users may run into problems deploying pods until you define your own policies. The recommend approach is to:

* Create an AKS cluster
* Define your own pod security policies
* Enable the pod security policy feature

To view the policies available, use the kubectl get psp command, as shown in the following example. As part of the default restricted policy, the user is denied PRIV use for privileged pod escalation, and the user MustRunAsNonRoot:
```bash
kubectl get psp
```

### Create a test user in an AKS cluster
Create a sample namespace named `psp-aks` for test resources using the `kubectl create namespace` command. Then, create a service account named `nonadmin-user` using the `kubectl create serviceaccount` command:
```bash
kubectl create namespace psp-aks
kubectl create serviceaccount --namespace psp-aks nonadmin-user
```

Next, create a RoleBinding for the nonadmin-user to perform basic actions in the namespace using the `kubectl create rolebinding` command:
```bash
kubectl create rolebinding \
    --namespace psp-aks \
    psp-aks-editor \
    --clusterrole=edit \
    --serviceaccount=psp-aks:nonadmin-user
```

To highlight the difference between the regular admin user when using kubectl and the non-admin user created in the previous steps, create two command-line aliases:
* The kubectl-admin alias is for the regular admin user, and is scoped to the psp-aks namespace.
* The kubectl-nonadminuser alias is for the nonadmin-user created in the previous step, and is scoped to the psp-aks namespace.

Create these two aliases as shown in the following commands:
```bash
alias kubectl-admin='kubectl --namespace psp-aks'
alias kubectl-nonadminuser='kubectl --as=system:serviceaccount:psp-aks:nonadmin-user --namespace psp-aks'
```

### Test the creation of a privileged pod

Let's first test what happens when you schedule a pod with the security context of privileged: true. This security context escalates the pod's privileges. In the previous section that showed the default AKS pod security policies, the restricted policy should deny this request.

Create the pod using the `kubectl apply` command and specify the name of your YAML manifest `nginx-privileged.yaml`:

```bash
kubectl-nonadminuser apply -f nginx-privileged.yaml
```
The pod fails to be scheduled and doesn't reach the scheduling stage.

### Test the creation of an unprivileged pod

In the previous example, the pod specification requested privileged escalation. This request is denied by the default restricted pod security policy, so the pod fails to be scheduled. Let's try now running that same NGINX pod without the privilege escalation request.

Create the pod using the `kubectl apply` command and specify the name of your YAML manifest `nginx-unprivileged.yaml`:

```bash
kubectl-nonadminuser apply -f nginx-unprivileged.yaml
```
The Kubernetes scheduler accepts the pod request. However, if you look at the status of the pod using kubectl get pods, there's an error:
```bash
kubectl-nonadminuser get pods
```

Use the kubectl describe pod command to look at the events for the pod. The following condensed example shows the container and image require root permissions, even though we didn't request them:
```bash
kubectl-nonadminuser describe pod nginx-unprivileged
```

Even though we didn't request any privileged access, the container image for NGINX needs to create a binding for port 80. To bind ports 1024 and below, the root user is required. When the pod tries to start, the restricted pod security policy denies this request.

This example shows that the default pod security policies created by AKS are in effect and restrict the actions a user can perform. It's important to understand the behavior of these default policies, as you may not expect a basic NGINX pod to be denied.

Before you move on to the next step, delete this test pod using the kubectl delete pod command:
```bash
kubectl-nonadminuser delete -f nginx-unprivileged.yaml
```

### Test creation of a pod with a specific user context
In the previous example, the container image automatically tried to use root to bind NGINX to port 80. This request was denied by the default restricted pod security policy, so the pod fails to start. Let's try now running that same NGINX pod with a specific user context, such as `runAsUser: 2000`.

Create the pod using the `nginx-unprivileged-nonroot.yaml` manifest:
```bash
kubectl-nonadminuser apply -f nginx-unprivileged-nonroot.yaml
```
he Kubernetes scheduler accepts the pod request. However, if you look at the status of the pod using kubectl get pods, there's a different error than the previous example:
```bash
kubectl-nonadminuser get pods
```

Use the kubectl describe pod command to look at the events for the pod. The following condensed example shows the pod events:
```bash
kubectl-nonadminuser describe pod nginx-unprivileged
```

The events indicate that the container was created and started. There's nothing immediately obvious as to why the pod is in a failed state. Let's look at the pod logs using the kubectl logs command:
```bash
kubectl-nonadminuser logs nginx-unprivileged-nonroot --previous
```
The following example log output gives an indication that within the NGINX configuration itself, there's a permissions error when the service tries to start. This error is again caused by needing to bind to port 80. Although the pod specification defined a regular user account, this user account isn't sufficient in the OS-level for the NGINX service to start and bind to the restricted port.

It's important to understand the behavior of the default pod security policies. This error was a little harder to track down, and again, you may not expect a basic NGINX pod to be denied.

Before you move on to the next step, delete this test pod using the kubectl delete pod command:
```bash
kubectl-nonadminuser delete -f nginx-unprivileged-nonroot.yaml
```

### Create a custom pod security policy

Now that you've seen the behavior of the default pod security policies, let's provide a way for the nonadmin-user to successfully schedule pods.

Let's create a policy to reject pods that request privileged access. Other options, such as runAsUser or allowed volumes, aren't explicitly restricted. This type of policy denies a request for privileged access, but otherwise lets the cluster run the requested pods.

Create the policy using the `kubectl apply` command and specify the name of provided YAML manifest:
```bash
kubectl apply -f psp-deny-privileged.yaml
```

### Allow user account tu use the custom pod security policy

In the previous step, you created a pod security policy to reject pods that request privileged access. To allow the policy to be used, you create a Role or a ClusterRole. Then, you associate one of these roles using a RoleBinding or ClusterRoleBinding.

For this example, create a ClusterRole that allows you to use the psp-deny-privileged policy created in the previous step.

Create the ClusterRole using the `kubectl apply` command and specify the name of provided YAML manifest:
```bash
kubectl apply -f psp-deny-privileged-clusterrole.yaml
```

Now create a ClusterRoleBinding to use the ClusterRole created in the previous step.

Create the ClusterRole using the `kubectl apply` command and specify the name of provided YAML manifest:
```bash
kubectl apply -f psp-deny-privileged-clusterrolebinding.yaml
```

### Test the creation of an unprivileged pod again

With your custom pod security policy applied and a binding for the user account to use the policy, let's try to create an unprivileged pod again. Use the same `nginx-privileged.yaml` manifest to create the pod using the `kubectl apply` command:

```bash
kubectl-nonadminuser apply -f nginx-unprivileged.yaml
```

The pod is successfully scheduled. When you check the status of the pod using the `kubectl get pods` command, the pod is Running:
```bash
kubectl-nonadminuser get pods
```
This example shows how you can create custom pod security policies to define access to the AKS cluster for different users or groups. The default AKS policies provide tight controls on what pods can run, so create your own custom policies to then correctly define the restrictions you need.

Delete the NGINX unprivileged pod using the kubectl delete command and specify the name of your YAML manifest:
```bash
kubectl-nonadminuser delete -f nginx-unprivileged.yaml
```