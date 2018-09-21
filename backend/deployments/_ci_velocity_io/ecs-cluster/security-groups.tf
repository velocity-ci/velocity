resource "aws_security_group" "instance_sg" {
  description = "Controls access to nodes in ECS cluster"
  vpc_id      = "${aws_vpc.app.id}"
  name        = "${var.cluster_name}.ecs-sg"

  ingress {
    protocol    = "-1"
    from_port   = 0
    to_port     = 0
    cidr_blocks = ["${var.cidr_block}"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags {
    cluster = "${var.cluster_name}"
    Name    = "${var.cluster_name}.ecs.instance"
  }
}
