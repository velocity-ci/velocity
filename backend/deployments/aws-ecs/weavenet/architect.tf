data "template_file" "ecs_def_architect" {
  template = "${file("${path.module}/architect.def.tpl.json")}"

  vars {
    version        = "${var.velocity_version}"
    jwt_secret     = "${var.jwt_secret}"
    builder_secret = "${var.builder_secret}"
    admin_password = "${var.admin_password}"

    architect_labels = "${jsonencode(var.architect_labels)}"

    logs_group  = "${var.cluster_name}.velocityci-container-logs"
    logs_region = "${var.aws_region}"

    weave_cidr = "${var.weave_cidr}"
  }
}

resource "aws_ecs_task_definition" "architect" {
  family                = "velocity_architect"
  container_definitions = "${data.template_file.ecs_def_architect.rendered}"
}

resource "aws_ecs_service" "architect" {
  name                               = "velocity_architect"
  cluster                            = "${var.cluster_name}"
  task_definition                    = "${aws_ecs_task_definition.architect.arn}"
  desired_count                      = 1
  deployment_minimum_healthy_percent = 100

  placement_strategy {
    type  = "spread"
    field = "attribute:ecs.availability-zone"
  }
}
