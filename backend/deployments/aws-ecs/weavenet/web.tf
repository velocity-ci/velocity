data "template_file" "ecs_def_web" {
  template = "${file("${path.module}/web.def.tpl.json")}"

  vars {
    version            = "${var.velocity_version}"
    architect_endpoint = "${var.architect_base_address}/v1"

    web_labels = "${jsonencode(var.web_labels)}"

    logs_group  = "${var.cluster_name}.velocityci-container-logs"
    logs_region = "${var.aws_region}"

    weave_cidr = "${var.weave_cidr}"
  }
}

resource "aws_ecs_task_definition" "web" {
  family                = "velocity_web"
  container_definitions = "${data.template_file.ecs_def_web.rendered}"
}

resource "aws_ecs_service" "web" {
  name                               = "velocity_web"
  cluster                            = "${var.cluster_name}"
  task_definition                    = "${aws_ecs_task_definition.web.arn}"
  desired_count                      = 1
  deployment_minimum_healthy_percent = 100

  placement_strategy {
    type  = "spread"
    field = "attribute:ecs.availability-zone"
  }
}
