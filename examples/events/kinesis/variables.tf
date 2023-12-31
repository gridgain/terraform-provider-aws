# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

variable "aws_region" {
  description = "The AWS region to create resources in."
  default     = "us-east-1"
}

variable "rule_name" {
  description = "The name of the EventBridge Rule"
  default     = "tf-example-cloudwatch-event-rule-for-kinesis"
}

variable "iam_role_name" {
  description = "The name of the IAM Role"
  default     = "tf-example-iam-role-for-kinesis"
}

variable "target_name" {
  description = "The name of the EventBridge Target"
  default     = "tf-example-cloudwatch-event-target-for-kinesis"
}

variable "stream_name" {
  description = "The name of the Kinesis Stream to send events to"
  default     = "tf-example-kinesis-stream"
}
