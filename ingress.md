# Creating Ingress

## Prerequisites:

1. Azure Subscription
2. Completed steps from [Docker Image Creation, Azure Container Registry, Azure Container Instances](https://github.com/akamenev/aks-workshop/blob/master/docker-images-acr-aci.md) and [Creating Pods, Deployments, Services](https://github.com/akamenev/aks-workshop/blob/master/creating-pods-deployments-services.md)
3. Helm - [installation guide](https://helm.sh/docs/using_helm/#installing-helm)
4. [Azure CLI installed](https://docs.microsoft.com/en-us/cli/azure/install-azure-cli?view=azure-cli-latest)

## Create welcome-app deployment and service
Inside the cloned repo go to `ingress` folder open `welcome-app-deployment.yaml` and replace REGISTRY_NAME with your registry name or use the command `sed` command below:
```yaml
...
      containers:
      - name: welcome-app
        image: REGISTRY_NAME.azurecr.io/welcome-app:v1
        ports:
        - containerPort: 8080
        resources:
          requests:
            cpu: 100m
            memory: 128Mi
          limits:
            cpu: 150m
            memory: 128Mi
      imagePullSecrets:
      - name: REGISTRY_NAME
```

```bash
sed -i "s/REGISTRY_NAME/$REGISTRY_NAME/g" welcome-app-deployment.yaml
```
Create welcome-app deployment and service
```bash
kubectl apply -f welcome-app-deployment.yaml
kubectl apply -f welcome-app-service-ingress.yaml
```

## Initialize helm
Inside the cloned repo go to `helm-demo` folder and run:
```
kubectl apply -f helm-rbac.yaml
helm init --service-account tiller
kubectl get pods -n kube-system | grep tiller
```

## Deploy the ingress controller with Helm

We will leverage the nip.io reverse wildcard DNS resolver service to map our ingress controller LoadBalancerIP to a proper DNS name.

```bash
helm repo update
helm upgrade --install ingress stable/nginx-ingress --namespace ingress
```

In a couple of minutes, a public IP address will be allocated to the ingress controller, retrieve with:
```bash
kubectl get svc  -n ingress    ingress-nginx-ingress-controller -o jsonpath="{.status.loadBalancer.ingress[*].ip}"

export INGRESS_IP=$(kubectl get svc  -n ingress    ingress-nginx-ingress-controller -o jsonpath="{.status.loadBalancer.ingress[*].ip}")
```

Open `welcome-ingress.yaml` and `welcome-ingress-tls.yaml` to replace INGRESS_IP with your IP or use the `sed` command below:
```yaml
...
spec:
  rules:
  - host: INGRESS_IP.nip.io
    http:
      paths:
...
```
```bash
sed -i "s/INGRESS_IP/$INGRESS_IP/g" welcome-ingress.yaml
```

```yaml
...
  tls:
  - hosts:
    - INGRESS_IP.nip.io
    secretName: welcome-tls-secret
  rules:
  - host: INGRESS_IP.nip.io
    http:
...
```
```bash
sed -i "s/INGRESS_IP/$INGRESS_IP/g" welcome-ingress-tls.yaml
```

Create an Ingress using `kubectl`:
```bash
kubectl apply -f welcome-ingress.yaml
```
Once the Ingress is deployed, you should be able to access the welcome app at http://INGRESS_IP.nip.io, for example http://52.255.217.198.nip.io

## Emable SSL/TLS on Ingress

You want to enable connecting to the welcome app over SSL/TLS. In this task, you’ll use Let’s Encrypt free service to generate valid SSL certificates for your domains, and you’ll integrate the certificate issuance workflow into Kubernetes.

Install `cert-manager`
```bash
helm install stable/cert-manager --name cert-manager --set ingressShim.defaultIssuerName=letsencrypt --set ingressShim.defaultIssuerKind=ClusterIssuer --version v0.5.2
```

Create a Let's Encrypt ClusterIssuer
Replace the EMAIL placeholder with your email in `letsencrypt-clusterissuer.yaml` or use the `sed` command below:
```yaml
...
    server: https://acme-v02.api.letsencrypt.org/directory # production
    #server: https://acme-staging-v02.api.letsencrypt.org/directory # staging
    email: EMAIL # replace this with your email
    privateKeySecretRef:
      name: letsencrypt
    http01: {}
...
```
```bash
sed -i "s/EMAIL/$EMAIL/g" letsencrypt-clusterissuer.yaml
```
And apply it:
```bash
kubectl apply -f letsencrypt-clusterissuer.yaml
```

Deploy the `welcome-ingress-tls.yaml`:
```bash
kubectl apply -f welcome-ingress-tls.yaml
```

Go to the app at `https://INGRESS_IP.nip.io`