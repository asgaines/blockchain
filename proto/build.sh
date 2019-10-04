#!/bin/sh

set -e

src=$(dirname $0)

go_output=./protogo

echo "$(date) test compiling for go"
prototool compile ${src}

echo "$(date) clearing old go output"
rm -rf ${go_output}/**/*.{pb,pb.gw}.go

echo "$(date) generating go proto"
prototool generate ${src}

echo "$(date) building go mocks.."
build_go_mocks(){
    echo "--> Building mock $2"
    mkdir -p $(dirname $3)
    # if we just pipe this to the destination directory bash creates the file
    # before mockgen starts attempting to reflect over the files, and it causes
    # mockgen to explode because it can't parse the output file
    mockgen $1 $2 | sed "s%github.com/asgaines/blockchain/vendor/%%" > /tmp/mock.go
    mv /tmp/mock.go $3
}