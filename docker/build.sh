#!/bin/sh
if [ $# -ne 1 ]; then
	echo "$0 <tag>"
	exit 0
fi

tag=$1

echo "tag: ${tag}"

rm -rf n9e pub
cp ../n9e .
cp -r ../pub .

docker build -t nightingale:${tag} .

docker tag nightingale:${tag} ccr.ccs.tencentyun.com/lightwan_ops/nightingale:latest
docker push ccr.ccs.tencentyun.com/lightwan_ops/nightingale:latest

docker tag nightingale:${tag} ccr.ccs.tencentyun.com/lightwan_ops/nightingale:${tag}
docker push ccr.ccs.tencentyun.com/lightwan_ops/nightingale:${tag}

rm -rf n9e pub
