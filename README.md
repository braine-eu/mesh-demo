Env
===
```
export KUBECONFIG=~/kubeconfig.yml
export CTX_CLUSTER1=cluster1
export CTX_CLUSTER2=cluster2
```

Install Istio 1.7.4
===
```
istioctl install --context $CTX_CLUSTER1 --set hub=braine-docker-local.artifactory.eng.vmware.com -f manifests/examples/multicluster/values-istio-multicluster-gateways.yaml
istioctl install --context $CTX_CLUSTER2 --set hub=braine-docker-local.artifactory.eng.vmware.com -f manifests/examples/multicluster/values-istio-multicluster-gateways.yaml
```

Setup DNS
===
```
scripts/setupdns.sh
```

Deploy demo
===
```
kubectl create --context=$CTX_CLUSTER1 namespace x86
kubectl label --context=$CTX_CLUSTER1 namespace x86 istio-injection=enabled
kubectl create --context=$CTX_CLUSTER2 namespace gpu
kubectl label --context=$CTX_CLUSTER2 namespace gpu istio-injection=enabled

kubectl apply --context=$CTX_CLUSTER1 -n x86 -f scripts/braine.yaml
kubectl apply --context=$CTX_CLUSTER1 -n x86 -f scripts/collect.yaml
kubectl apply --context=$CTX_CLUSTER2 -n gpu -f scripts/plot.yaml
scripts/mksvcentry.sh
kubectl apply --context=$CTX_CLUSTER1 -n x86 -f svcentry.yaml
```

Clean up
===
```
kubectl delete --context=$CTX_CLUSTER1 -n x86 -f scripts/braine.yaml
kubectl delete --context=$CTX_CLUSTER1 -n x86 -f scripts/collect.yaml
kubectl delete --context=$CTX_CLUSTER2 -n gpu -f scripts/plot.yaml
kubectl delete --context=$CTX_CLUSTER1 -n x86 -f svcentry.yaml
```
