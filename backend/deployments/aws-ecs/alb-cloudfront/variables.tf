variable "velocity_version" {
  type = "string"
}

variable "region" {
  type = "string"
}

variable "domain" {
  type = "string"
}

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

variable "debug" {
  default = "false"
}
