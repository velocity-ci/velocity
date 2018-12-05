resource "aws_cloudwatch_log_group" "velocityci" {
  name = "${var.cluster_name}.velocityci-container-logs"

  retention_in_days = 7

  tags {
    Name = "Velocity CI"
  }
}
