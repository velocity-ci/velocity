#!/bin/bash

set -e

export VM_NAME=velocity
echo "Setting VM_IP"
export VM_IP=$(hostname -i)
echo "Set VM_IP as ${LOCAL_IP}"


echo "Waiting for database to become available"
/bin/wait-for-it.sh -t 120 ${DATABASE_HOSTNAME}:${DATABASE_PORT}

/opt/app/bin/velocity $@
