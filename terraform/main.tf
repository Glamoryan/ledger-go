provider "aws" {
  region = var.region
}

terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.0"
    }
  }
  required_version = ">= 1.0.0"
}

# Lambda fonksiyonu için IAM rolü
resource "aws_iam_role" "lambda_role" {
  name = "${var.app_name}-lambda-role-${var.environment}"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "lambda.amazonaws.com"
        }
      }
    ]
  })
  
  tags = {
    Environment = var.environment
    Name        = "${var.app_name}-lambda-role-${var.environment}"
  }
}

# Lambda için izinler (SQS ve CloudWatch Logs için)
resource "aws_iam_policy" "lambda_policy" {
  name        = "${var.app_name}-lambda-policy-${var.environment}"
  description = "Lambda için gerekli izinler"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "sqs:ReceiveMessage",
          "sqs:DeleteMessage",
          "sqs:GetQueueAttributes",
          "sqs:SendMessage"
        ]
        Effect   = "Allow"
        Resource = [aws_sqs_queue.ledger_queue.arn, aws_sqs_queue.dlq.arn]
      },
      {
        Action = [
          "logs:CreateLogGroup",
          "logs:CreateLogStream",
          "logs:PutLogEvents"
        ]
        Effect   = "Allow"
        Resource = "arn:aws:logs:*:*:*"
      },
      {
        Action = [
          "ec2:CreateNetworkInterface",
          "ec2:DeleteNetworkInterface",
          "ec2:DescribeNetworkInterfaces"
        ]
        Effect   = "Allow"
        Resource = "*"
      }
    ]
  })
  
  tags = {
    Environment = var.environment
    Name        = "${var.app_name}-lambda-policy-${var.environment}"
  }
}

# Lambda rolüne izinleri ata
resource "aws_iam_role_policy_attachment" "lambda_policy_attach" {
  role       = aws_iam_role.lambda_role.name
  policy_arn = aws_iam_policy.lambda_policy.arn
}

# SQS Dead Letter Queue (DLQ)
resource "aws_sqs_queue" "dlq" {
  name                       = "${var.app_name}-dlq-${var.environment}"
  delay_seconds              = 0
  max_message_size           = 262144
  message_retention_seconds  = 604800 # 7 days
  visibility_timeout_seconds = 60
  receive_wait_time_seconds  = 10

  tags = {
    Environment = var.environment
    Name        = "${var.app_name}-dlq-${var.environment}"
  }
}

# SQS kuyruğu
resource "aws_sqs_queue" "ledger_queue" {
  name                       = "${var.app_name}-queue-${var.environment}"
  delay_seconds              = 0
  max_message_size           = 262144
  message_retention_seconds  = 345600 # 4 days
  visibility_timeout_seconds = 60
  receive_wait_time_seconds  = 10
  
  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.dlq.arn
    maxReceiveCount     = 3
  })

  tags = {
    Environment = var.environment
    Name        = "${var.app_name}-queue-${var.environment}"
  }
}

# CloudWatch Logs grubu
resource "aws_cloudwatch_log_group" "lambda_logs" {
  name              = "/aws/lambda/${var.app_name}-processor-${var.environment}"
  retention_in_days = 7
  
  tags = {
    Environment = var.environment
    Name        = "${var.app_name}-lambda-logs-${var.environment}"
  }
}

# Ana lambda fonksiyonu
resource "aws_lambda_function" "ledger_processor" {
  filename          = var.lambda_zip_path
  function_name     = "${var.app_name}-processor-${var.environment}"
  role              = aws_iam_role.lambda_role.arn
  handler           = "bootstrap"
  source_code_hash  = filebase64sha256(var.lambda_zip_path)
  runtime           = "provided.al2"
  timeout           = 30
  memory_size       = 128

  vpc_config {
    subnet_ids         = aws_subnet.private[*].id
    security_group_ids = [aws_security_group.lambda_sg.id]
  }

  environment {
    variables = {
      ENVIRONMENT    = var.environment
      SQS_QUEUE_URL  = aws_sqs_queue.ledger_queue.url
      DB_HOST        = aws_db_instance.ledger_db.address
      DB_PORT        = tostring(aws_db_instance.ledger_db.port)
      DB_NAME        = var.db_name
      DB_USER        = var.db_username
      DB_PASSWORD    = var.db_password
    }
  }

  depends_on = [
    aws_iam_role_policy_attachment.lambda_policy_attach,
    aws_cloudwatch_log_group.lambda_logs,
    aws_db_instance.ledger_db
  ]
}

# SQS'den Lambda'yı tetikleme
resource "aws_lambda_event_source_mapping" "sqs_trigger" {
  event_source_arn = aws_sqs_queue.ledger_queue.arn
  function_name    = aws_lambda_function.ledger_processor.arn
  batch_size       = 5
  enabled          = true
}

# API Gateway için IAM rolü (SQS'e mesaj gönderebilmesi için)
resource "aws_iam_role" "api_gateway_sqs" {
  name = "${var.app_name}-api-gateway-sqs-role-${var.environment}"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "apigateway.amazonaws.com"
        }
      }
    ]
  })
  
  tags = {
    Environment = var.environment
    Name        = "${var.app_name}-api-gateway-sqs-role-${var.environment}"
  }
}

# API Gateway'in SQS'e mesaj gönderebilmesi için politika
resource "aws_iam_policy" "api_gateway_sqs_policy" {
  name        = "${var.app_name}-api-gateway-sqs-policy-${var.environment}"
  description = "API Gateway'in SQS kuyruğuna mesaj göndermesi için izin"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "sqs:SendMessage"
        ]
        Effect   = "Allow"
        Resource = [aws_sqs_queue.ledger_queue.arn]
      }
    ]
  })
  
  tags = {
    Environment = var.environment
    Name        = "${var.app_name}-api-gateway-sqs-policy-${var.environment}"
  }
}

# API Gateway rolüne SQS politikasını ekle
resource "aws_iam_role_policy_attachment" "api_gateway_sqs_attach" {
  role       = aws_iam_role.api_gateway_sqs.name
  policy_arn = aws_iam_policy.api_gateway_sqs_policy.arn
}
