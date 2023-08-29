#!/bin/bash

set -e

LIBSSL11_AMD64_DL_URL="http://security.ubuntu.com/ubuntu/pool/main/o/openssl/libssl1.1_1.1.1-1ubuntu2.1~18.04.23_amd64.deb"
LIBSSL_DEV_11_AMD64_DL_URL="http://security.ubuntu.com/ubuntu/pool/main/o/openssl/libssl-dev_1.1.1-1ubuntu2.1~18.04.23_amd64.deb"
LIBSSL11_ARM64_DL_URL="http://ports.ubuntu.com/pool/main/o/openssl/libssl1.1_1.1.1f-1ubuntu2.19_arm64.deb"
LIBSSL11_DEV_ARM64_DL_URL="http://ports.ubuntu.com/pool/main/o/openssl/libssl-dev_1.1.1f-1ubuntu2.19_arm64.deb"

arch=$(dpkg --print-architecture)

if [[ $arch == "arm64" ]]; then
  curl $LIBSSL11_ARM64_DL_URL -o libssl1.1.deb
  curl $LIBSSL11_DEV_ARM64_DL_URL -o libssl-dev_1.1.deb
else
  curl $LIBSSL11_AMD64_DL_URL -o libssl1.1.deb
  curl $LIBSSL_DEV_11_AMD64_DL_URL -o libssl-dev_1.1.deb
fi

dpkg -i libssl1.1.deb
dpkg -i libssl-dev_1.1.deb

# copy to common place for next stage
target=$(uname -m)
cp -R /usr/lib/$target-linux-gnu/engines-1.1 /usr/lib/engines-1.1
cp /usr/lib/$target-linux-gnu/libcrypto.so.1.1 /usr/lib/libcrypto.so.1.1
cp /usr/lib/$target-linux-gnu/libssl.so.1.1 /usr/lib/libssl.so.1.1
