#!/bin/bash -e

cwd=$(pwd)

cd "${cwd}/../backend/"

scripts/build-cli.sh

cd "${cwd}"

install_path=$(pwd)
bin_path="${install_path}/vcli"

if [[ "${OSTYPE}" == "linux-gnu" ]]; then
    arch="linux"
elif [[ "${OSTYPE}" == "darwin"* ]]; then
    arch="darwin"
else
    echo "-> unsupported OS: ${OSTYPE}"
    exit 1
fi
echo "-> determined ${arch} OS"
cp "${cwd}/../backend/dist/vcli_${arch}_amd64" "${bin_path}"
echo "copied ${cwd}/../backend/dist/vcli_${arch}_amd64 to ${bin_path}"
