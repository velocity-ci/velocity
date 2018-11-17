variable "velocity_version" {
  type = "string"
}

variable "architect_base_address" {
  type = "string"
}

variable "jwt_secret" {
  type = "string"
}

variable "builder_secret" {
  type = "string"
}

variable "admin_password" {
  type    = "string"
  default = ""
}

variable "aws_region" {
  type = "string"
}

variable "cluster_name" {
  type = "string"
}

variable "weave_cidr" {
  type = "string"
}

variable "provision_alb" {
  type        = "string"
  description = "set to 'true' to provision an ALB for the architect"
  default     = ""
}

variable "architect_labels" {
  type    = "map"
  default = {}
}

variable "web_labels" {
  type    = "map"
  default = {}
}
