#!/bin/bash -e

if [[ "${OSTYPE}" == "linux-gnu" ]]; then
    arch="linux"
elif [[ "${OSTYPE}" == "darwin"* ]]; then
    arch="darwin"
else
    echo "-> unsupported OS: ${OSTYPE}"
    exit 1
fi
echo "-> determined ${arch} OS"

install_path=$(pwd)

bin_path="${install_path}/vcli"
echo "-> downloading to ${bin_path}"
repo="velocity-ci/velocity"
download_url=$(curl -s https://api.github.com/repos/${repo}/releases/latest \
  | grep browser_download_url \
  | grep ${arch} \
  | cut -d '"' -f 4)
echo "-> downloading ${download_url} to ${bin_path}"
curl -L $download_url -o ${bin_path}
chmod +x ${bin_path}
