#!/bin/bash
export INGRESS_HOST=$(kubectl --context=$CTX_CLUSTER2 get po -l istio=ingressgateway -n istio-system -o jsonpath='{.items[0].status.hostIP}')
export INGRESS_PORT=$(kubectl --context=$CTX_CLUSTER2 get svc -n istio-system istio-ingressgateway -o=jsonpath='{.spec.ports[?(@.port==15443)].nodePort}')

echo "Ingress host for $CTX_CLUSTER2: $INGRESS_HOST"
echo "Ingress port for $CTX_CLUSTER2: $INGRESS_PORT"
cat <<EOF > svcentry.yaml
apiVersion: networking.istio.io/v1alpha3
kind: ServiceEntry
metadata:
  name: plot-gpu
spec:
  hosts:
  # must be of form name.namespace.global
  - plot.gpu.global
  # Treat remote cluster services as part of the service mesh
  # as all clusters in the service mesh share the same root of trust.
  location: MESH_INTERNAL
  ports:
  - name: http2
    number: 80
    protocol: http
  resolution: DNS
  addresses:
  # the IP address to which httpbin.bar.global will resolve to
  # must be unique for each remote service, within a given cluster.
  # This address need not be routable. Traffic for this IP will be captured
  # by the sidecar and routed appropriately.
  - 240.0.0.3
  endpoints:
  # This is the routable address of the ingress gateway in cluster2 that
  # sits in front of sleep.foo service. Traffic from the sidecar will be
  # routed to this address.
  - address: ${INGRESS_HOST}
    ports:
      http2: ${INGRESS_PORT}
EOF
echo "Service entry script saved to svcentry.yaml"
