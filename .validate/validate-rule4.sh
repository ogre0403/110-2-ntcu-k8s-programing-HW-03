#!/bin/bash

echo ""
echo "測試刪除nginx Deployment後，informer 建立的 Service也被刪除"


deployment=nginx-deployment
kubectl delete deployments.apps ${deployment} >/dev/null  2>&1

LABEL="ntcu-k8s=hw3"


for i in {1..20}; do
  sleep 1
  svc_num=`kubectl get svc   -l ${LABEL}  -o yaml | yq '.items | length'`
  if [[ "$svc_num" -eq 0 ]]; then
      break
  fi

  if [[ "$i" -eq 20 ]]; then
      echo "timeout: informer 刪除建立的svc, 應為0,  $svc_num 不正確"
      exit 1
  fi

done

echo "........ PASS"
