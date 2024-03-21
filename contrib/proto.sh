#!/bin/bash

set -e

decho(){
    1>&2 echo $@
}

PROTO_BIN="protoc"

PROTO_DIR="./contrib/proto"
PROTO_GO_DIR="./proto"

BASEURL="github.com"
REPO="${BASEURL}/noncepad/solana-tx-processor"


test_all(){
    decho $@
}

build_go(){
    rm -r $PROTO_GO_DIR 2>>/dev/null || true
    $PROTO_BIN --experimental_allow_proto3_optional --proto_path=${PROTO_DIR} --go-grpc_out=. --go_out=. $@
    #proto/basic.proto proto/serum.proto proto/solana-net.proto
    mv "${REPO}/proto" ./
    rm -r "${BASEURL}"
}



build_go $(echo $(find ./contrib/proto -type f))

#build_python $(echo $(find proto -type f))

#build_cpp $(echo $(find proto -type f))

#build_csharp $(echo $(find proto -type f))
