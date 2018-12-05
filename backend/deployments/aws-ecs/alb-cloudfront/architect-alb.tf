resource "aws_route53_record" "architect" {
  zone_id = "${data.aws_route53_zone.organisation.zone_id}"
  name    = "architect.${var.domain}"
  type    = "A"

  alias {
    name                   = "${aws_alb.architect.dns_name}"
    zone_id                = "${aws_alb.architect.zone_id}"
    evaluate_target_health = false
  }
}

resource "aws_alb" "architect" {
  name            = "architect-velocityci-alb"
  subnets         = ["${data.aws_subnet_ids.app_cluster.ids}"]
  security_groups = ["${aws_security_group.alb_sg.id}"]

  provisioner "local-exec" {
    command = "sleep 10"
  }
}

resource "aws_alb_target_group" "architect" {
  name                 = "architect-velocityci-tg"
  port                 = 80
  protocol             = "HTTP"
  vpc_id               = "${data.aws_vpc.app_cluster.id}"
  deregistration_delay = "30"

  health_check {
    path                = "/v1/health"
    healthy_threshold   = 2
    unhealthy_threshold = 2
    timeout             = 2
    protocol            = "HTTP"
    interval            = 10
    matcher             = "200"
  }
}

resource "aws_security_group" "alb_sg" {
  description = "Controls access to and from the ALB"

  vpc_id = "${data.aws_vpc.app_cluster.id}"
  name   = "architect.velocityci.alb-sg"

  ingress {
    protocol    = "tcp"
    from_port   = 80
    to_port     = 80
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    protocol    = "tcp"
    from_port   = 443
    to_port     = 443
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port = 0
    to_port   = 0
    protocol  = "-1"

    cidr_blocks = [
      "0.0.0.0/0",
    ]
  }
}

resource "aws_iam_role" "ecs_service" {
  name = "architect.velocityci.ecs"

  assume_role_policy = <<EOF
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "ecs.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "ecs_service" {
  name = "architect.velocityci.ecs"
  role = "${aws_iam_role.ecs_service.name}"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ec2:Describe*",
        "elasticloadbalancing:DeregisterInstancesFromLoadBalancer",
        "elasticloadbalancing:DeregisterTargets",
        "elasticloadbalancing:Describe*",
        "elasticloadbalancing:RegisterInstancesWithLoadBalancer",
        "elasticloadbalancing:RegisterTargets"
      ],
      "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_alb_listener" "architect" {
  load_balancer_arn = "${aws_alb.architect.id}"
  port              = "80"
  protocol          = "HTTP"

  default_action {
    target_group_arn = "${aws_alb_target_group.architect.id}"
    type             = "forward"
  }
}

resource "aws_alb_listener" "front_end_ssl" {
  load_balancer_arn = "${aws_alb.architect.id}"
  port              = "443"
  protocol          = "HTTPS"
  ssl_policy        = "ELBSecurityPolicy-2016-08"
  certificate_arn   = "${data.aws_acm_certificate.architect.arn}"

  default_action {
    target_group_arn = "${aws_alb_target_group.architect.id}"
    type             = "forward"
  }
}

data "aws_acm_certificate" "architect" {
  domain   = "architect.${var.domain}"
  statuses = ["ISSUED"]
}
