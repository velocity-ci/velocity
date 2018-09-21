terraform {
  backend "s3" {
    encrypt = true
    bucket  = "velocityci-tfstate"
    key     = "cluster/terraform.tfstate"
    region  = "eu-west-1"
  }
}

provider "template" {
  version = "~> 1.0"
}

provider "aws" {
  version = "~> 1.32"
  region  = "${var.region}"
}

## AWS ECS - Elastic Container Service
resource "aws_ecs_cluster" "app" {
  name = "${var.cluster_name}"
}
