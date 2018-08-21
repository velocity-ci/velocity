data "template_file" "user_data" {
  template = "${file("${path.module}/user-data.sh")}"

  vars {
    ecs_cluster_name  = "${aws_ecs_cluster.app.name}"
    ecs_log_level     = "info"
    ecs_agent_version = "latest"
  }
}

resource "aws_iam_instance_profile" "app" {
  name = "${var.cluster_name}.ecs"
  role = "${aws_iam_role.app_instance.name}"
}

resource "aws_iam_role_policy" "app_instance" {
  name = "${var.cluster_name}.app_policy"
  role = "${aws_iam_role.app_instance.name}"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ec2:*",
        "ecs:*",
        "autoscaling:DescribeAutoScalingInstances",
        "ecr:GetAuthorizationToken",
        "ecr:BatchCheckLayerAvailability",
        "ecr:GetDownloadUrlForLayer",
        "ecr:BatchGetImage",
        "logs:CreateLogStream",
        "logs:PutLogEvents",
        "route53:*",
        "route53domains:*"
      ],
      "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_iam_role" "app_instance" {
  name = "${var.cluster_name}.ecs-instance-role"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "ec2.amazonaws.com"
        ]
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}
