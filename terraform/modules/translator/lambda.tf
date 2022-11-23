resource "aws_s3_object" "app" {
  bucket = var.bucket_id

  key    = "${local.app_name}/app.zip"
  source = var.archive_file.output_path
  etag   = filemd5(var.archive_file.output_path)
}

resource "aws_iam_role" "app_lambda_exec" {
  name = local.app_name

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Sid    = ""
      Principal = {
        Service = "lambda.amazonaws.com"
      }
      }
    ]
  })
}

data "aws_iam_policy_document" "app" {
  // allow running `aws sts get-caller-identity`
  statement {
    effect = "Allow"
    actions = [
      "translate:*",
      "secretsmanager:GetResourcePolicy",
      "secretsmanager:GetSecretValue",
      "secretsmanager:DescribeSecret",
      "secretsmanager:ListSecretVersionIds",
      "kms:Decrypt"
    ]
    resources = ["*"]
  }
}

resource "aws_iam_policy" "app" {
  name        = local.app_name
  path        = "/"
  description = "Policy for ${local.app_name}"
  policy      = data.aws_iam_policy_document.app.json
}

resource "aws_iam_role_policy_attachment" "app_lambda_policy" {
  role       = aws_iam_role.app_lambda_exec.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_iam_role_policy_attachment" "app_policy" {
  role       = aws_iam_role.app_lambda_exec.name
  policy_arn = aws_iam_policy.app.arn
}

resource "aws_lambda_function" "app" {
  function_name = local.app_name
  s3_bucket     = var.bucket_id
  s3_key        = aws_s3_object.app.key

  runtime = "go1.x"
  handler = "main"

  source_code_hash = var.archive_file.output_base64sha256

  role = aws_iam_role.app_lambda_exec.arn

  memory_size = 128
  timeout     = 30
}

resource "aws_cloudwatch_log_group" "translator" {
  name = "/aws/lambda/${aws_lambda_function.app.function_name}"

  retention_in_days = 30
}