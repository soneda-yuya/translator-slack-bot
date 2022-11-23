variable "bucket_id" {
  description = "slack bot bucket"

  type = string
}

variable "archive_file" {
  description = "slack bot bucket"

  type = object({
    source_dir : string
    output_path : string
    output_base64sha256 : string
  })
}

locals {
  app_name = "slack-translator"
}