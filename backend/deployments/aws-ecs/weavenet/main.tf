data "template_file" "ecs_def_architect" {
  template = "${file("${path.module}/architect.def.tpl.json")}"

  vars {
    version        = "${var.version}"
    jwt_secret     = "${var.jwt_secret}"
    builder_secret = "${var.builder_secret}"
    admin_password = "${var.admin_password}"

    cloudwatch_log_group = "${aws_cloudwatch_log_group.api.arn}"
    cloudwatch_region    = "${var.aws_region}"

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
  iam_role                           = "${aws_iam_role.ecs_service.arn}"
  deployment_minimum_healthy_percent = 100

  placement_strategy {
    type  = "spread"
    field = "attribute:ecs.availability-zone"
  }
}
