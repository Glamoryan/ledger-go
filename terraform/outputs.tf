output "api_url" {
  description = "API Gateway endpoint URL"
  value       = "${aws_api_gateway_deployment.api_deployment.invoke_url}${aws_api_gateway_stage.api_stage.stage_name}"
}

output "api_key" {
  description = "API Key"
  value       = aws_api_gateway_api_key.api_key.value
  sensitive   = true
}

output "sqs_queue_url" {
  description = "SQS queue URL"
  value       = aws_sqs_queue.ledger_queue.url
}

output "sqs_dlq_url" {
  description = "SQS dead letter queue URL"
  value       = aws_sqs_queue.dlq.url
}

output "lambda_function_name" {
  description = "Lambda function name"
  value       = aws_lambda_function.ledger_processor.function_name
}

output "rds_endpoint" {
  description = "RDS endpoint"
  value       = aws_db_instance.ledger_db.address
}

output "rds_port" {
  description = "RDS port"
  value       = aws_db_instance.ledger_db.port
}

output "rds_database_name" {
  description = "RDS database name"
  value       = aws_db_instance.ledger_db.db_name
}

output "vpc_id" {
  description = "VPC ID"
  value       = aws_vpc.main.id
} 