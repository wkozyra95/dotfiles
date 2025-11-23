# Setup

Go to `k8s` directory

Create cluster:
```
kind create cluster --config=cluster.yaml
```

Configure tailscale secrets:
```
kubectl create namespace tailscale
kubectl create secret generic operator-oauth --namespace tailscale --from-literal="client_id=<CLIENT_ID>" --from-literal="client_secret=<CLIENT_SECRET>"
```

Ensure `/storage` exists:
```
mkdir -p /storage && chown wojtek:users /storage
```

Deploy services:
```
kubectl kustomize . --enable-helm | kubectl apply -f -
```


## Show dashboard

On k8s host
```
kubectl port-forward --namespace kubernetes-dashboard services/dashboard-kong-proxy  8000:80
```

On main device
```
ssh -L 8000:127.0.0.1:8000 wojtek@192.168.100.5
```

Open `localhost:8000` in the browser

Token can be generated with:
```
kubectl --namespace kubernetes-dashboard create token dashboard-admin-ro
```
