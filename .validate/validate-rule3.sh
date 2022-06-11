#!/bin/bash

echo ""
echo "驗證Informer 是否會自動建立Service，並測試是否連連通"


LABEL="ntcu-k8s=hw3"

# informer 會建立一個svc

for i in {1..20}; do
  sleep 1
  svc_num=`kubectl get svc   -l ${LABEL}  -o yaml | yq '.items | length'`

  if [[ "$svc_num" -eq 1 ]]; then
      break
  fi

  if [[ "$i" -eq 20 ]]; then
      echo "timeout: informer建立的svc數量 $svc_num 不正確. 應為 1"
      exit 1
  fi

done

cid=`docker ps -f name=control-plane -q`


nodeport=`kubectl get svc  -l ${LABEL}  -o jsonpath='{.items[0].spec.ports[0].nodePort}'`

for i in {1..20}; do

  docker exec ${cid} curl 127.0.0.1:${nodeport}  >/dev/null  2>&1

  RET=$?

  if [[ ${RET} -eq 0 ]]; then
    echo "........ PASS"
    exit 0
  fi

  sleep 3
done

echo "timeout for wait for connect deployment"
exit 1

