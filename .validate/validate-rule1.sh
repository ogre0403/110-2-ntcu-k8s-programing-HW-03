#!/bin/bash

echo ""
echo "驗證是否有建立Client Deployment，以及 使用的ServiceAccount 權限"
##
##

sa=`for f in ./manifest/*; do cat ${f} | yq '(.|select(.kind == "ServiceAccount")).metadata.name' ; done`

deployment=`for f in ./manifest/*; do cat ${f} | yq "(.|select(.spec.template.spec.serviceAccountName == \"${sa}\")).metadata.name" ; done`



kubectl get deployment $deployment >/dev/null  2>&1

RET=$?


if [[ $RET != 0 ]]; then
    echo "查無apps/v1 Deployment: $deployment"
    exit 1
fi


kubectl get sa $sa >/dev/null  2>&1

RET=$?


if [[ $RET != 0 ]]; then
    echo "查無core/v1 ServiceAccount: $sa"
    exit 1
fi

cani=`kubectl auth can-i get    svc --as system:serviceaccount:default:$sa`
if [[ $cani == "no" ]]; then
    echo "ServiceAccount $sa 無法查詢svc"
    exit 1
fi


cani=`kubectl auth can-i create svc --as system:serviceaccount:default:$sa`
if [[ $cani == "no" ]]; then
    echo "ServiceAccount $sa 無法建立svc"
    exit 1
fi

cani=`kubectl auth can-i delete svc --as system:serviceaccount:default:$sa`
if [[ $cani == "no" ]]; then
    echo "ServiceAccount $sa 無法刪除svc"
    exit 1
fi

cani=`kubectl auth can-i get    deployment --as system:serviceaccount:default:$sa`
if [[ $cani == "no" ]]; then
    echo "ServiceAccount $sa 無法查詢deployment"
    exit 1
fi

cani=`kubectl auth can-i create deployment --as system:serviceaccount:default:$sa`
if [[ $cani == "no" ]]; then
    echo "ServiceAccount $sa 無法建立deployment"
    exit 1
fi


cani=`kubectl auth can-i delete deployment --as system:serviceaccount:default:$sa`
if [[ $cani == "no" ]]; then
    echo "ServiceAccount $sa 無法刪除deployment"
    exit 1
fi


echo "........ PASS"