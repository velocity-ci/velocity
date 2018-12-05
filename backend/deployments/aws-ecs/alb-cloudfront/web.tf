data "template_file" "ecs_def_web" {
  template = "${file("${path.module}/web.def.tpl.json")}"

  vars {
    version           = "${var.velocity_version}"
    architect_address = "https://${aws_route53_record.architect.name}/v1"

    logs_group  = "${var.cluster_name}.velocityci-container-logs"
    logs_region = "${data.aws_region.current.name}"
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

  load_balancer {
    target_group_arn = "${aws_alb_target_group.web.arn}"
    container_name   = "velocityci_web"
    container_port   = 80
  }

  ordered_placement_strategy {
    type  = "spread"
    field = "attribute:ecs.availability-zone"
  }
}
