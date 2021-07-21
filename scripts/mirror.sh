#!/bin/bash
echo "Starting"
SOURCE_HUB=istio
DEST_HUB="braine-docker-local.artifactory.eng.vmware.com"
IMAGES="install-cni operator pilot proxyv2" # Images to mirror.
VERSIONS="1.7.4 1.7.5" # Versions to copy
for image in $IMAGES; do
for version in $VERSIONS; do
  name=$image:$version
  #echo $name
  docker pull $SOURCE_HUB/$name
  docker tag $SOURCE_HUB/$name $DEST_HUB/$name
  docker push $DEST_HUB/$name
  docker rmi $SOURCE_HUB/$name
  docker rmi $DEST_HUB/$name

  name=$image:$version-distroless
  #echo $name
  docker pull $SOURCE_HUB/$name
  docker tag $SOURCE_HUB/$name $DEST_HUB/$name
  docker push $DEST_HUB/$name
  docker rmi $SOURCE_HUB/$name
  docker rmi $DEST_HUB/$name
done
done

# one off for istio-coredns
name="coredns-plugin:0.2-istio-1.1"
docker pull $SOURCE_HUB/$name
docker tag $SOURCE_HUB/$name $DEST_HUB/$name
docker push $DEST_HUB/$name
docker rmi $SOURCE_HUB/$name
docker rmi $DEST_HUB/$name

echo "Done."
