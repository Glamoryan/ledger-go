# API Gateway tanımı
resource "aws_api_gateway_rest_api" "ledger_api" {
  name        = "${var.app_name}-api-${var.environment}"
  description = "Ledger API Gateway"
  
  endpoint_configuration {
    types = ["REGIONAL"]
  }
  
  tags = {
    Environment = var.environment
    Name        = "${var.app_name}-api-${var.environment}"
  }
}

# API Gateway için kök kaynak (/) ve metod
resource "aws_api_gateway_method" "root_method" {
  rest_api_id   = aws_api_gateway_rest_api.ledger_api.id
  resource_id   = aws_api_gateway_rest_api.ledger_api.root_resource_id
  http_method   = "GET"
  authorization = "NONE"
  api_key_required = true
}

resource "aws_api_gateway_integration" "root_integration" {
  rest_api_id             = aws_api_gateway_rest_api.ledger_api.id
  resource_id             = aws_api_gateway_rest_api.ledger_api.root_resource_id
  http_method             = aws_api_gateway_method.root_method.http_method
  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = aws_lambda_function.ledger_processor.invoke_arn
}

# Users resource
resource "aws_api_gateway_resource" "users" {
  rest_api_id = aws_api_gateway_rest_api.ledger_api.id
  parent_id   = aws_api_gateway_rest_api.ledger_api.root_resource_id
  path_part   = "users"
}

# Users add-user sub-resource
resource "aws_api_gateway_resource" "add_user" {
  rest_api_id = aws_api_gateway_rest_api.ledger_api.id
  parent_id   = aws_api_gateway_resource.users.id
  path_part   = "add-user"
}

# Add User POST method
resource "aws_api_gateway_method" "add_user_post" {
  rest_api_id   = aws_api_gateway_rest_api.ledger_api.id
  resource_id   = aws_api_gateway_resource.add_user.id
  http_method   = "POST"
  authorization = "NONE"
  api_key_required = true
}

# Add User integration with SQS
resource "aws_api_gateway_integration" "add_user_sqs_integration" {
  rest_api_id             = aws_api_gateway_rest_api.ledger_api.id
  resource_id             = aws_api_gateway_resource.add_user.id
  http_method             = aws_api_gateway_method.add_user_post.http_method
  integration_http_method = "POST"
  type                    = "AWS"
  uri                     = "arn:aws:apigateway:${var.region}:sqs:path/${aws_sqs_queue.ledger_queue.name}"
  credentials             = aws_iam_role.api_gateway_sqs.arn
  
  request_parameters = {
    "integration.request.header.Content-Type" = "'application/x-www-form-urlencoded'"
  }

  request_templates = {
    "application/json" = <<EOF
Action=SendMessage
&MessageBody=$util.urlEncode("{\"path\": \"/users/add-user\", \"httpMethod\": \"POST\", \"body\": $input.json('$'), \"headers\": {#foreach($header in $input.params().header.keySet())\"$header\": \"$util.escapeJavaScript($input.params().header.get($header))\"#if($foreach.hasNext),#end#end}}")
EOF
  }
}

# Add User method response
resource "aws_api_gateway_method_response" "add_user_post_response_200" {
  rest_api_id = aws_api_gateway_rest_api.ledger_api.id
  resource_id = aws_api_gateway_resource.add_user.id
  http_method = aws_api_gateway_method.add_user_post.http_method
  status_code = "200"
  
  response_models = {
    "application/json" = "Empty"
  }
}

# Add User integration response
resource "aws_api_gateway_integration_response" "add_user_sqs_integration_response" {
  rest_api_id = aws_api_gateway_rest_api.ledger_api.id
  resource_id = aws_api_gateway_resource.add_user.id
  http_method = aws_api_gateway_method.add_user_post.http_method
  status_code = aws_api_gateway_method_response.add_user_post_response_200.status_code
  
  response_templates = {
    "application/json" = jsonencode({
      message = "Kullanıcı oluşturma isteği kuyruğa alındı",
      status  = "success"
    })
  }
  
  depends_on = [aws_api_gateway_integration.add_user_sqs_integration]
}

# Login resource
resource "aws_api_gateway_resource" "login" {
  rest_api_id = aws_api_gateway_rest_api.ledger_api.id
  parent_id   = aws_api_gateway_rest_api.ledger_api.root_resource_id
  path_part   = "login"
}

# Login POST method
resource "aws_api_gateway_method" "login_post" {
  rest_api_id   = aws_api_gateway_rest_api.ledger_api.id
  resource_id   = aws_api_gateway_resource.login.id
  http_method   = "POST"
  authorization = "NONE"
  api_key_required = true
}

# Login integration with Lambda proxy
resource "aws_api_gateway_integration" "login_lambda_integration" {
  rest_api_id             = aws_api_gateway_rest_api.ledger_api.id
  resource_id             = aws_api_gateway_resource.login.id
  http_method             = aws_api_gateway_method.login_post.http_method
  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = aws_lambda_function.ledger_processor.invoke_arn
}

# Register resource
resource "aws_api_gateway_resource" "register" {
  rest_api_id = aws_api_gateway_rest_api.ledger_api.id
  parent_id   = aws_api_gateway_rest_api.ledger_api.root_resource_id
  path_part   = "register"
}

# Register POST method - Yetkilendirme gerektirmeyen endpoint
resource "aws_api_gateway_method" "register_post" {
  rest_api_id   = aws_api_gateway_rest_api.ledger_api.id
  resource_id   = aws_api_gateway_resource.register.id
  http_method   = "POST"
  authorization = "NONE"
  api_key_required = true
}

# Register integration with Lambda proxy
resource "aws_api_gateway_integration" "register_lambda_integration" {
  rest_api_id             = aws_api_gateway_rest_api.ledger_api.id
  resource_id             = aws_api_gateway_resource.register.id
  http_method             = aws_api_gateway_method.register_post.http_method
  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = aws_lambda_function.ledger_processor.invoke_arn
}

# Users send-credit sub-resource
resource "aws_api_gateway_resource" "send_credit" {
  rest_api_id = aws_api_gateway_rest_api.ledger_api.id
  parent_id   = aws_api_gateway_resource.users.id
  path_part   = "send-credit"
}

# Send Credit POST method
resource "aws_api_gateway_method" "send_credit_post" {
  rest_api_id   = aws_api_gateway_rest_api.ledger_api.id
  resource_id   = aws_api_gateway_resource.send_credit.id
  http_method   = "POST"
  authorization = "NONE"
  api_key_required = true
}

# Send Credit integration with Lambda proxy
resource "aws_api_gateway_integration" "send_credit_lambda_integration" {
  rest_api_id             = aws_api_gateway_rest_api.ledger_api.id
  resource_id             = aws_api_gateway_resource.send_credit.id
  http_method             = aws_api_gateway_method.send_credit_post.http_method
  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = aws_lambda_function.ledger_processor.invoke_arn
}

# Users get-credit sub-resource
resource "aws_api_gateway_resource" "get_credit" {
  rest_api_id = aws_api_gateway_rest_api.ledger_api.id
  parent_id   = aws_api_gateway_resource.users.id
  path_part   = "get-credit"
}

# Get Credit GET method
resource "aws_api_gateway_method" "get_credit_get" {
  rest_api_id   = aws_api_gateway_rest_api.ledger_api.id
  resource_id   = aws_api_gateway_resource.get_credit.id
  http_method   = "GET"
  authorization = "NONE"
  api_key_required = true
}

# Get Credit integration with Lambda proxy
resource "aws_api_gateway_integration" "get_credit_lambda_integration" {
  rest_api_id             = aws_api_gateway_rest_api.ledger_api.id
  resource_id             = aws_api_gateway_resource.get_credit.id
  http_method             = aws_api_gateway_method.get_credit_get.http_method
  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = aws_lambda_function.ledger_processor.invoke_arn
}

# API Key tanımı
resource "aws_api_gateway_api_key" "api_key" {
  name        = "${var.app_name}-api-key-${var.environment}"
  description = "API Key for ${var.app_name}"
  enabled     = true
}

# API Key kullanım planı
resource "aws_api_gateway_usage_plan" "usage_plan" {
  name        = "${var.app_name}-usage-plan-${var.environment}"
  description = "${var.app_name} API kullanım planı"
  
  api_stages {
    api_id = aws_api_gateway_rest_api.ledger_api.id
    stage  = aws_api_gateway_stage.api_stage.stage_name
  }
  
  quota_settings {
    limit  = 1000
    period = "DAY"
  }
  
  throttle_settings {
    burst_limit = 5
    rate_limit  = 10
  }
  
  depends_on = [
    aws_api_gateway_deployment.api_deployment,
    aws_api_gateway_stage.api_stage
  ]
}

# API Key ile kullanım planı ilişkilendirme
resource "aws_api_gateway_usage_plan_key" "usage_plan_key" {
  key_id        = aws_api_gateway_api_key.api_key.id
  key_type      = "API_KEY"
  usage_plan_id = aws_api_gateway_usage_plan.usage_plan.id
}

# API Gateway dağıtımı
resource "aws_api_gateway_deployment" "api_deployment" {
  rest_api_id = aws_api_gateway_rest_api.ledger_api.id
  
  triggers = {
    redeployment = sha1(jsonencode([
      aws_api_gateway_resource.users.id,
      aws_api_gateway_resource.add_user.id,
      aws_api_gateway_resource.login.id,
      aws_api_gateway_resource.register.id,
      aws_api_gateway_resource.get_credit.id,
      aws_api_gateway_resource.send_credit.id,
      aws_api_gateway_method.root_method.id,
      aws_api_gateway_method.add_user_post.id,
      aws_api_gateway_method.login_post.id,
      aws_api_gateway_method.register_post.id,
      aws_api_gateway_method.get_credit_get.id,
      aws_api_gateway_method.send_credit_post.id,
      aws_api_gateway_integration.root_integration.id,
      aws_api_gateway_integration.add_user_sqs_integration.id,
      aws_api_gateway_integration.login_lambda_integration.id,
      aws_api_gateway_integration.register_lambda_integration.id,
      aws_api_gateway_integration.get_credit_lambda_integration.id,
      aws_api_gateway_integration.send_credit_lambda_integration.id
    ]))
  }
  
  lifecycle {
    create_before_destroy = true
  }
  
  depends_on = [
    aws_api_gateway_method.root_method,
    aws_api_gateway_integration.root_integration,
    aws_api_gateway_method.add_user_post,
    aws_api_gateway_integration.add_user_sqs_integration,
    aws_api_gateway_method.login_post,
    aws_api_gateway_integration.login_lambda_integration,
    aws_api_gateway_method.register_post,
    aws_api_gateway_integration.register_lambda_integration,
    aws_api_gateway_method.get_credit_get,
    aws_api_gateway_integration.get_credit_lambda_integration,
    aws_api_gateway_method.send_credit_post,
    aws_api_gateway_integration.send_credit_lambda_integration
  ]
}

# API Gateway stage
resource "aws_api_gateway_stage" "api_stage" {
  deployment_id = aws_api_gateway_deployment.api_deployment.id
  rest_api_id   = aws_api_gateway_rest_api.ledger_api.id
  stage_name    = var.environment
}

# Lambda'ya API Gateway'dan çağrı izni
resource "aws_lambda_permission" "api_gw_lambda" {
  statement_id  = "AllowExecutionFromAPIGateway"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.ledger_processor.function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_api_gateway_rest_api.ledger_api.execution_arn}/*/*"
} 