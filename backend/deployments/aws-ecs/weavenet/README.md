## Velocity CI Deployment

### AWS ECS w/ WeaveNet

#### w/ Traefik
```
provider "aws" {
  region = "eu-west-1"
  version = "~> 1.11"
}

provider "template" {
  version = "~> 1.0"
}

terraform {
  backend "s3" {
    encrypt = true
    bucket  = "org-terraform-state"
    key     = "ci-velocity/terraform.tfstate"
    region  = "eu-west-1"
  }
}

data "aws_route53_zone" "organisation" {
  name = "example.org."
}

resource "aws_route53_record" "architect" {
  zone_id = "${data.aws_route53_zone.organisation.zone_id}"
  name    = "architect.velocity.example.org"
  type    = "CNAME"
  ttl     = "300"
  records = ["traefik.example.org"]
}

resource "aws_route53_record" "web" {
  zone_id = "${data.aws_route53_zone.organisation.zone_id}"
  name    = "velocity.example.org"
  type    = "CNAME"
  ttl     = "300"
  records = ["traefik.example.org"]
}

module "velocityci" {
    source = "github.com/velocity-ci/velocity//backend/deployments/aws-ecs/weavenet"
    aws_region = "eu-west-1"

    velocity_version = "807849d"

    cluster_name = "org"
    weave_cidr = "10.32.105.0/24"

    jwt_secret = "org-test"
    builder_secret = "org-test"
    admin_password = "org-test"

    provision_alb = "false"

    architect_base_address = "https://architect.velocity.example.org"

    architect_labels = {
        "traefik.frontend.rule" = "Host:architect.velocity.example.org",
        "traefik.enable" = "true",
        "traefik.frontend.entrypoints" = "https,wss"
    }

    web_labels = {
         "traefik.frontend.rule" = "Host:velocity.example.org",
        "traefik.enable" = "true",
        "traefik.frontend.entrypoints" = "https,wss"
    }
}
```


#### w/ ALB