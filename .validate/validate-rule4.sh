#!/bin/bash

echo ""
echo "測試刪除nginx Deployment後，informer 建立的 Service也被刪除"

#sa=`for f in ./manifest/*; do cat ${f} | yq '(.|select(.kind == "ServiceAccount")).metadata.name' ; done`
#deployment=`for f in ./manifest/*; do cat ${f} | yq "(.|select(.spec.template.spec.serviceAccountName == \"${sa}\")).metadata.name" ; done`

deployment=nginx-deployment
kubectl delete deployments.apps ${deployment} >/dev/null  2>&1

LABEL="ntcu-k8s=hw3"

svc_num=`kubectl get svc   -l ${LABEL}  -o yaml | yq '.items | length'`


if [[ "$svc_num" -ne 0 ]]; then
    echo "informer 刪除建立的svc, 應為0,  $svc_num 不正確"
    exit 1
fi

#deployment_num=`kubectl get deployment -l ${LABEL}  -o yaml | yq '.items | length'`
#if [[ "$deployment_num" -ne 0 ]]; then
#    echo "client建立的deployment 數量 $deployment_num 不正確"
#    exit 1
#fi

echo "........ PASS"
