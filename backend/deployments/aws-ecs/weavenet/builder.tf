data "template_file" "ecs_def_builder" {
  template = "${file("${path.module}/builder.def.tpl.json")}"

  vars {
    version        = "${var.velocity_version}"
    builder_secret = "${var.builder_secret}"

    logs_group  = "${var.cluster_name}.velocityci-container-logs"
    logs_region = "${var.aws_region}"

    weave_cidr = "${var.weave_cidr}"
  }
}

resource "aws_ecs_task_definition" "builder" {
  family                = "velocity_builder"
  container_definitions = "${data.template_file.ecs_def_builder.rendered}"

  task_role_arn = "${aws_iam_role.builder.arn}"

   volume {
    name      = "docker-engine"
    host_path = "/var/run/docker.sock"
  }

  volume {
    name = "velocity-workspace"
    host_path = "/opt/velocityci"
  }
}

resource "aws_ecs_service" "builder" {
  name                               = "velocity_builder"
  cluster                            = "${var.cluster_name}"
  task_definition                    = "${aws_ecs_task_definition.builder.arn}"
  desired_count                      = 1
  deployment_minimum_healthy_percent = 100

  placement_strategy {
    type  = "spread"
    field = "attribute:ecs.availability-zone"
  }
}

resource "aws_iam_role" "builder" {
  name = "ecs.velocityci.builder"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AllowECSTasksAssumeRole",
      "Effect": "Allow",
      "Principal": {
        "Service": "ecs-tasks.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "builder" {
  role = "${aws_iam_role.builder.name}"
  name = "ecs.velocityci.builder"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "*",
      "Effect": "Allow",
      "Resource": "*"
    }
  ]
}
EOF
}