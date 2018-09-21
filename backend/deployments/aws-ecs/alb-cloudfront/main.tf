provider "aws" {
  region  = "eu-west-1"
  version = "~> 1.23"
}

provider "template" {
  version = "~> 1.0"
}

data "aws_region" "current" {}
