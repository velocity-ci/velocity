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
repo="velocity-ci/velocity"
download_url=$(curl -s https://api.github.com/repos/${repo}/releases/latest \
  | grep browser_download_url \
  | grep ${arch} \
  | cut -d '"' -f 4)

bin_folder="${HOME}/.local/bin"

mkdir -p ${bin_folder}
bin_path="${bin_folder}/vcli"
echo "-> downloading ${download_url}"
curl -L $download_url -o ${bin_path}
chmod +x ${bin_path}

[[ ":$PATH:" != *":${bin_folder}:"* ]] && PATH="/path/to/add:${PATH}"

vcli -h
