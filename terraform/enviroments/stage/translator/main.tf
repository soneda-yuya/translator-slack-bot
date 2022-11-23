terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "4.40.0"
    }
    archive = {
      source  = "hashicorp/archive"
      version = "~> 2.2.0"
    }
  }

  backend "s3" {
    bucket = "your-slack-bot-stage"
    key    = "tfstate/terraform.tfstate"
    region = "ap-northeast-1"
  }
}

provider "aws" {
  region = "ap-northeast-1"
}

resource "aws_s3_bucket" "slack_bot" {
  bucket = "your-slack-bot-stage"
}

resource "aws_s3_bucket_acl" "slack_bot" {
  bucket = aws_s3_bucket.slack_bot.id
  acl    = "private"
}

resource "aws_s3_bucket_versioning" "slack_bot" {
  bucket = aws_s3_bucket.slack_bot.id
  versioning_configuration {
    status = "Disabled"
  }
}

data "archive_file" "app" {
  type = "zip"

  source_dir  = "${path.module}/../../../../lambdas/translator/build"
  output_path = "${path.module}/app.zip"
}

module "app" {
  source    = "../../../modules/translator"
  bucket_id = aws_s3_bucket.slack_bot.id
  archive_file = {
    source_dir : data.archive_file.app.source_dir
    output_path : data.archive_file.app.output_path
    output_base64sha256 : data.archive_file.app.output_base64sha256
  }
}
