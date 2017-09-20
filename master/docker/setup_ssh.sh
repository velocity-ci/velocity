#!/bin/sh

for host in $(echo ${SSH_KNOWN_HOSTS} | sed "s/,/ /g")
do
    echo "Adding ${host} to known_hosts"
    ssh-keyscan ${host} >> ~/.ssh/known_hosts
done

eval $(ssh-agent)
