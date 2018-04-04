variable "cluster_name" {
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

variable "provision_alb" {
  type        = "string"
  description = "set to 'true' to provision an ALB for the architect"
  default     = ""
}

variable "architect_labels" {
  type = "map"
}
