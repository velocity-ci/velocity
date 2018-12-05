data "aws_availability_zones" "available" {}

resource "aws_vpc" "app" {
  cidr_block = "${var.cidr_block}"

  enable_dns_support   = true
  enable_dns_hostnames = true

  tags {
    cluster = "${var.cluster_name}"
    Name    = "${var.cluster_name}.ecs"
  }
}

resource "aws_subnet" "app" {
  count             = "${length(data.aws_availability_zones.available.names)}"
  cidr_block        = "${cidrsubnet(aws_vpc.app.cidr_block, var.az_size, count.index)}"
  availability_zone = "${data.aws_availability_zones.available.names[count.index]}"
  vpc_id            = "${aws_vpc.app.id}"

  tags {
    cluster = "${var.cluster_name}"
    Name    = "${var.cluster_name}.ecs.${data.aws_availability_zones.available.names[count.index]}"
  }
}

resource "aws_internet_gateway" "app" {
  vpc_id = "${aws_vpc.app.id}"
}

resource "aws_route_table" "app" {
  vpc_id = "${aws_vpc.app.id}"

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = "${aws_internet_gateway.app.id}"
  }

  tags {
    cluster = "${var.cluster_name}"
    Name    = "${var.cluster_name}.ecs"
  }
}

resource "aws_route_table_association" "a" {
  count          = "${length(data.aws_availability_zones.available.names)}"
  subnet_id      = "${element(aws_subnet.app.*.id, count.index)}"
  route_table_id = "${aws_route_table.app.id}"
}
