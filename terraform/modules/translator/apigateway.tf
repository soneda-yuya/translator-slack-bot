resource "aws_apigatewayv2_api" "app_lambda" {
  name          = "${local.app_name}_gw"
  protocol_type = "HTTP"
}

resource "aws_apigatewayv2_stage" "app_lambda" {
  api_id = aws_apigatewayv2_api.app_lambda.id

  name        = local.app_name
  auto_deploy = true

  access_log_settings {
    destination_arn = aws_cloudwatch_log_group.api_gw.arn

    format = jsonencode({
      requestId               = "$context.requestId"
      sourceIp                = "$context.identity.sourceIp"
      requestTime             = "$context.requestTime"
      protocol                = "$context.protocol"
      httpMethod              = "$context.httpMethod"
      resourcePath            = "$context.resourcePath"
      routeKey                = "$context.routeKey"
      status                  = "$context.status"
      responseLength          = "$context.responseLength"
      integrationErrorMessage = "$context.integrationErrorMessage"
      }
    )
  }
}

resource "aws_apigatewayv2_integration" "app" {
  api_id = aws_apigatewayv2_api.app_lambda.id

  integration_uri    = aws_lambda_function.app.invoke_arn
  integration_type   = "AWS_PROXY"
  integration_method = "POST"
}

resource "aws_apigatewayv2_route" "app" {
  api_id = aws_apigatewayv2_api.app_lambda.id

  route_key = "POST /exec"
  target    = "integrations/${aws_apigatewayv2_integration.app.id}"
}

resource "aws_cloudwatch_log_group" "api_gw" {
  name = "/aws/api_gw/${aws_apigatewayv2_api.app_lambda.name}"

  retention_in_days = 30
}

resource "aws_lambda_permission" "api_gw" {
  statement_id  = "AllowExecutionFromAPIGateway"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.app.function_name
  principal     = "apigateway.amazonaws.com"

  source_arn = "${aws_apigatewayv2_api.app_lambda.execution_arn}/*/*"
}