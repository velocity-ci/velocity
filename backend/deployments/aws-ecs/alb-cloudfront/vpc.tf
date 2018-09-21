data "aws_vpc" "app_cluster" {
  filter {
    name   = "tag:cluster"
    values = ["${var.cluster_name}"]
  }
}

data "aws_subnet_ids" "app_cluster" {
  vpc_id = "${data.aws_vpc.app_cluster.id}"
}
