#!/bin/bash

# Copyright 2019 Istio Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -e
set -x

WD=$(dirname "$0")
WD=$(cd "$WD"; pwd)
ROOT=$(dirname "$WD")
cd "${ROOT}"

ISTIO_DIR="${GOPATH}/src/istio.io/istio"
if [[ ! -d "${ISTIO_DIR}" ]]
then
    ISTIO_DIR=$(mktemp -d)/src/istio.io/istio
    git clone https://github.com/istio/istio.git "${ISTIO_DIR}"
fi
ISTIO_CONTROL_NS=istio-system
MODE=permissive
SIMPLE_AUTH=false
E2E_ARGS="--skip_setup=true --use_local_cluster=true --istio_namespace=${ISTIO_CONTROL_NS}"
TMPDIR=/tmp

export GO111MODULE=on
export IstioTop=${ISTIO_DIR}/../../..

#kind and istioctl setup
KIND_IMAGE="kindest/node:v1.14.0"
pushd ${ISTIO_DIR}
go install ./istioctl/cmd/istioctl
export ISTIOCTL_BIN=${GOPATH}/bin/istioctl
source "./prow/lib.sh"
setup_kind_cluster ${KIND_IMAGE}
popd

echo "installing istio with operator CLI"
go run ./cmd/mesh.go manifest apply

function run-simple-base() {
    kubectl create ns ${NS} || true
    kubectl -n ${NS} apply -f prow/k8s/mtls_${MODE}.yaml
    kubectl -n ${NS} apply -f prow/k8s/sidecar-local.yaml
    kubectl label ns ${NS} istio-injection=disabled --overwrite
    (cd ${ISTIO_DIR}; make e2e_simple_run ${TEST_FLAGS} \
    E2E_ARGS="${E2E_ARGS} --auth_enable=${SIMPLE_AUTH} --namespace=${NS}")
}
function run-simple() {
    ISTIO_CONTROL_NS=istio-system MODE=permissive NS=simple run-simple-base
}
# Simple test, strict mode
function run-simple-strict() {
    MODE=strict ISTIO_CONTROL_NS=istio-system NS=simple-strict SIMPLE_AUTH=true run-simple-base
}

function run-bookinfo-demo() {
    kubectl create ns bookinfo-demo || true
    kubectl -n bookinfo-demo apply -f prow/k8s/mtls_permissive.yaml
    kubectl -n bookinfo-demo apply -f prow/k8s/sidecar-local.yaml
    (cd ${ISTIO_DIR}; make e2e_bookinfo_run ${TEST_FLAGS} \
      E2E_ARGS="${E2E_ARGS} --namespace=bookinfo-demo")
}

echo "start e2e testing"
run-simple
run-simple-strict
run-bookinfo-demo


