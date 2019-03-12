#!/usr/bin/env sh

PROJ=cluster-manager

TAG=${1:-develop}
PROPELLER_TAG=${2:-develop}

HASH=`date +%s`
BUILD_NAME=${PROJ}-build-${HASH}

rm -rf target
mkdir target
docker build --no-cache  -t ${BUILD_NAME} --build-arg BB_API_KEY=$BB_API_KEY --build-arg PROPELLER_TAG=$PROPELLER_TAG -f Dockerfile.build . || exit 1
docker create --name ${BUILD_NAME} ${BUILD_NAME} /bin/true || exit 1
docker cp ${BUILD_NAME}:/target/$PROJ ./target/$PROJ || exit 1
docker rm ${BUILD_NAME} || exit 1
docker rmi -f ${BUILD_NAME} || exit 1

docker build --no-cache \
    --build-arg KOPS_VERSION=1.7.1 \
    --build-arg KUBECTL_VERSION=$(wget -qO- https://storage.googleapis.com/kubernetes-release/release/stable.txt) \
    -t applariat/${PROJ}:${TAG} . || exit 1
rm -rf target
