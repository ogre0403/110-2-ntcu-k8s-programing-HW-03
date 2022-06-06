#!/bin/bash

echo ""
echo "驗證Client Deployment 是否有建立web service，並測試是否連連通"

LABEL="ntcu-k8s=hw2"

svc_num=`kubectl get svc   -l ${LABEL}  -o yaml | yq '.items | length'`


if [[ "$svc_num" -ne 1 ]]; then
    echo "client建立的svc數量 $svc_num 不正確"
    exit 1
fi

deployment_num=`kubectl get deployment -l ${LABEL}  -o yaml | yq '.items | length'`
if [[ "$deployment_num" -ne 1 ]]; then
    echo "client建立的deployment 數量 $deployment_num 不正確"
    exit 1
fi

# find out nginx deployment name
deployment=`kubectl get deployment -l ${LABEL}  -o yaml | yq '.items[0].metadata.name'`

ready="false"
for i in {1..60}; do
  ready=`kubectl get deployments.apps ${deployment} >/dev/null  2>&1  && kubectl get deployments.apps ${deployment} -o yaml |  yq .status.readyReplicas==.status.replicas`

	if [ "$ready" == "true" ]; then
		break
	fi
    echo "deployment ${deployment} not ready, sleep 1 sec"
done



cid=`docker ps -f name=control-plane -q`


nodeport=`kubectl get svc  -l ntcu-k8s=hw2  -o jsonpath='{.items[0].spec.ports[0].nodePort}'`

for i in {1..20}; do

  docker exec ${cid} curl 127.0.0.1:${nodeport}  >/dev/null  2>&1

  RET=$?

  if [[ ${RET} -eq 0 ]]; then
    echo "........ PASS"
    exit 0
  fi

  sleep 3
done

echo "timeout for wait for connect deployment ${deployment} success"
exit 1

