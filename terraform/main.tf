variable "identifier" {}

resource "aws_iam_user" "rds-operator-k8s" {
  name = "${format("%s-rds-operator", var.identifier)}"
}

resource "aws_iam_role" "rds-operator" {
  name_prefix = "${format("%s-rds-operator", var.identifier)}"
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
  "AWS": "${aws_iam_user.rds-operator-k8s.arn}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_policy" "rds-operator" {
  name_prefix = "${format("%s-rds-operator", var.identifier)}"
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "rds:*"
      ],
      "Effect": "Allow",
      "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "rds-operator" {
  role       = "${aws_iam_role.rds-operator.name}"
  policy_arn = "${aws_iam_policy.rds-operator.arn}"
}

output "user-arn" {
  value = "${aws_iam_user.rds-operator-k8s.arn}"
}

output "role-arn" {
  value = "${aws_iam_role.rds-operator.arn}"
}
