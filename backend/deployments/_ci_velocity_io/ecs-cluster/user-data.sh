#!/bin/bash -e
#
# For latest Ubuntu 16.04 LTS AMI (currently: ami-a8d2d7ce)

#
# Install Docker CE (from https://docs.docker.com/engine/installation/linux/ubuntu/#install-using-the-repository)
apt-get update

apt-get install -y \
    apt-transport-https \
    ca-certificates \
    curl \
    software-properties-common

curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -

add-apt-repository \
   "deb [arch=amd64] https://download.docker.com/linux/ubuntu \
   $(lsb_release -cs) \
   stable"

apt-get update
apt-get install -y docker-ce
docker run --rm hello-world


rm -rf /var/log/ecs/* /var/lib/ecs/data/*

#
# Install ECS Agent
sh -c "echo 'net.ipv4.conf.all.route_localnet = 1' >> /etc/sysctl.conf"
sysctl -p /etc/sysctl.conf
iptables -t nat -A PREROUTING -p tcp -d 169.254.170.2 --dport 80 -j DNAT --to-destination 127.0.0.1:51679
iptables -t nat -A OUTPUT -d 169.254.170.2 -p tcp -m tcp --dport 80 -j REDIRECT --to-ports 51679
sh -c 'iptables-save > /etc/network/iptables.rules'
mkdir -p /etc/ecs && touch /etc/ecs/ecs.config
cat <<EOF >> /etc/ecs/ecs.config
ECS_DATADIR=/data
ECS_ENABLE_TASK_IAM_ROLE=true
ECS_ENABLE_TASK_IAM_ROLE_NETWORK_HOST=true
ECS_LOGFILE=/log/ecs-agent.log
ECS_AVAILABLE_LOGGING_DRIVERS=["json-file","awslogs"]
ECS_LOGLEVEL=info
ECS_CLUSTER=${ecs_cluster_name}
EOF
echo ""
echo "Joined ECS Cluster: ${ecs_cluster_name}"
echo ""

mkdir -p /var/log/ecs /var/lib/ecs/data

cat <<EOF >> /etc/systemd/system/ecs.service
[Unit]
Description=Amazon ECS agent
Requires=docker.service
After=network.target
StartLimitInterval=200
StartLimitBurst=5
[Service]
Type=simple
ExecStartPre=-/usr/bin/docker stop ecs-agent
ExecStartPre=-/usr/bin/docker rm ecs-agent
ExecStartPre=/usr/bin/docker pull amazon/amazon-ecs-agent:latest
ExecStart=/usr/bin/docker run --name ecs-agent --restart=on-failure:10 --volume=/var/run:/var/run --volume=/var/log/ecs/:/log --volume=/var/lib/ecs/data:/data --volume=/etc/ecs:/etc/ecs --net=host --env-file=/etc/ecs/ecs.config amazon/amazon-ecs-agent:latest
Restart=always
RestartSec=30
[Install]
WantedBy=multi-user.target
EOF
chmod 755 /etc/systemd/system/ecs.service
systemctl daemon-reload
systemctl enable ecs
systemctl start ecs