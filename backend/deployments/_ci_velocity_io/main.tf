provider "aws" {
  region  = "eu-west-1"
  version = "~> 1.23"
}

terraform {
  backend "s3" {
    encrypt = true
    bucket  = "velocityci-tfstate"
    key     = "ci-velocity.tfstate"
    region  = "eu-west-1"
  }
}

data "aws_ecs_cluster" "org" {
  cluster_name = "${var.cluster_name}"
}

module "velocityci" {
  source = "../aws-ecs/alb-cloudfront"
  region = "${var.region}"

  domain           = "${var.domain}"
  velocity_version = "a7af32a"

  cluster_name = "${data.aws_ecs_cluster.org.cluster_name}"

  jwt_secret     = "${var.jwt_secret}"
  builder_secret = "${var.builder_secret}"
  admin_password = "${var.admin_password}"
}
