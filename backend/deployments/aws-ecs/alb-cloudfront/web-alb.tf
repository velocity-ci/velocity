resource "aws_route53_record" "web" {
  zone_id = "${data.aws_route53_zone.organisation.zone_id}"
  name    = "ci.${var.domain}"
  type    = "A"

  alias {
    name                   = "${aws_alb.web.dns_name}"
    zone_id                = "${aws_alb.web.zone_id}"
    evaluate_target_health = false
  }
}

resource "aws_alb" "web" {
  name            = "web-velocityci-alb"
  subnets         = ["${data.aws_subnet_ids.app_cluster.ids}"]
  security_groups = ["${aws_security_group.web_alb.id}"]

  provisioner "local-exec" {
    command = "sleep 10"
  }
}

resource "aws_alb_target_group" "web" {
  name                 = "web-velocityci-tg"
  port                 = 80
  protocol             = "HTTP"
  vpc_id               = "${data.aws_vpc.app_cluster.id}"
  deregistration_delay = "30"

  health_check {
    path                = "/"
    healthy_threshold   = 2
    unhealthy_threshold = 2
    timeout             = 2
    protocol            = "HTTP"
    interval            = 10
    matcher             = "200"
  }
}

resource "aws_security_group" "web_alb" {
  description = "Controls access to and from the ALB"

  vpc_id = "${data.aws_vpc.app_cluster.id}"
  name   = "web.velocityci.alb-sg"

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

resource "aws_iam_role" "web" {
  name = "web.velocityci.ecs"

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

resource "aws_iam_role_policy" "web" {
  name = "web.velocityci.ecs"
  role = "${aws_iam_role.web.name}"

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

resource "aws_alb_listener" "web" {
  load_balancer_arn = "${aws_alb.web.id}"
  port              = "80"
  protocol          = "HTTP"

  default_action {
    target_group_arn = "${aws_alb_target_group.web.id}"
    type             = "forward"
  }
}

resource "aws_alb_listener" "web_ssl" {
  load_balancer_arn = "${aws_alb.web.id}"
  port              = "443"
  protocol          = "HTTPS"
  ssl_policy        = "ELBSecurityPolicy-2016-08"
  certificate_arn   = "${data.aws_acm_certificate.web.arn}"

  default_action {
    target_group_arn = "${aws_alb_target_group.web.id}"
    type             = "forward"
  }
}

data "aws_acm_certificate" "web" {
  domain   = "ci.${var.domain}"
  statuses = ["ISSUED"]
}
