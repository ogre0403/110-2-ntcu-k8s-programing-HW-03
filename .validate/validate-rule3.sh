#!/bin/bash

echo ""
echo "測試Client Deployment週期讀取Web Deployment 名字至少四次"
LABEL="ntcu-k8s=hw2"


sa=`for f in ./manifest/*; do cat ${f} | yq '(.|select(.kind == "ServiceAccount")).metadata.name' ; done`


deploy_name=`kubectl get deployments.apps -l ${LABEL} -o yaml | yq '.items[0].metadata.name'`

client_pod=`kubectl get pod -o yaml | yq "(.items[]|select(.spec.serviceAccountName == \"${sa}\")).metadata.name"`




for i in {1..20}; do
  occur_count=`kubectl logs ${client_pod} | grep -o ${deploy_name} -c`
  if [[ "$occur_count" -ge 4 ]]; then
      echo "........ PASS"
      exit 0
  fi
  sleep 3
done


echo "timeout. 經過 60 秒，${deploy_name} 出現少於4次"
exit 1
