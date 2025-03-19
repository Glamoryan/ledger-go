# VPC ve Subnet tanımlamaları
resource "aws_vpc" "main" {
  cidr_block           = var.vpc_cidr
  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = {
    Name        = "${var.app_name}-vpc-${var.environment}"
    Environment = var.environment
  }
}

resource "aws_subnet" "private" {
  count             = 2
  vpc_id            = aws_vpc.main.id
  cidr_block        = cidrsubnet(var.vpc_cidr, 8, count.index)
  availability_zone = data.aws_availability_zones.available.names[count.index]

  tags = {
    Name        = "${var.app_name}-private-subnet-${count.index}-${var.environment}"
    Environment = var.environment
  }
}

# İnternet Gateway
resource "aws_internet_gateway" "igw" {
  vpc_id = aws_vpc.main.id

  tags = {
    Name        = "${var.app_name}-igw-${var.environment}"
    Environment = var.environment
  }
}

# Route Table
resource "aws_route_table" "private" {
  vpc_id = aws_vpc.main.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.igw.id
  }

  tags = {
    Name        = "${var.app_name}-private-route-${var.environment}"
    Environment = var.environment
  }
}

resource "aws_route_table_association" "private" {
  count          = 2
  subnet_id      = aws_subnet.private[count.index].id
  route_table_id = aws_route_table.private.id
}

# Security Group
resource "aws_security_group" "db_sg" {
  name        = "${var.app_name}-db-sg-${var.environment}"
  description = "Security group for RDS"
  vpc_id      = aws_vpc.main.id

  ingress {
    from_port   = 5432
    to_port     = 5432
    protocol    = "tcp"
    cidr_blocks = [var.vpc_cidr]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name        = "${var.app_name}-db-sg-${var.environment}"
    Environment = var.environment
  }
}

# Lambda için security group
resource "aws_security_group" "lambda_sg" {
  name        = "${var.app_name}-lambda-sg-${var.environment}"
  description = "Security group for Lambda to connect to RDS"
  vpc_id      = aws_vpc.main.id

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name        = "${var.app_name}-lambda-sg-${var.environment}"
    Environment = var.environment
  }
}

# DB Subnet Group
resource "aws_db_subnet_group" "default" {
  name       = "${var.app_name}-db-subnet-group-${var.environment}"
  subnet_ids = aws_subnet.private[*].id

  tags = {
    Name        = "${var.app_name}-db-subnet-group-${var.environment}"
    Environment = var.environment
  }
}

# RDS Veritabanı
resource "aws_db_instance" "ledger_db" {
  identifier              = "${var.app_name}-db-${var.environment}"
  engine                  = "postgres"
  engine_version          = "14.12"
  instance_class          = "db.t3.micro"
  allocated_storage       = 20
  storage_type            = "gp2"
  username                = var.db_username
  password                = var.db_password
  db_name                 = var.db_name
  vpc_security_group_ids  = [aws_security_group.db_sg.id]
  db_subnet_group_name    = aws_db_subnet_group.default.name
  skip_final_snapshot     = true
  publicly_accessible     = true  # Geliştirme için public erişim açık, üretim için false yapılmalı
  backup_retention_period = 7
  deletion_protection     = false # Geliştirme için false, üretim için true yapılmalı

  tags = {
    Name        = "${var.app_name}-db-${var.environment}"
    Environment = var.environment
  }
}

# Lambda için RDS politikası
resource "aws_iam_policy" "lambda_rds_policy" {
  name        = "${var.app_name}-lambda-rds-policy-${var.environment}"
  description = "Policy for Lambda to connect to RDS"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "rds-db:connect"
        ]
        Effect   = "Allow"
        Resource = aws_db_instance.ledger_db.arn
      }
    ]
  })
}

# Lambda rolüne RDS politikası ekle
resource "aws_iam_role_policy_attachment" "lambda_rds_attach" {
  role       = aws_iam_role.lambda_role.name
  policy_arn = aws_iam_policy.lambda_rds_policy.arn
}

# Availability Zones
data "aws_availability_zones" "available" {
  state = "available"
}

# RDS oluşturulduktan sonra veritabanı şemasını ve ilk verileri yükle
resource "null_resource" "db_setup" {
  count = 0  # Geçici olarak devre dışı bırak
  depends_on = [aws_db_instance.ledger_db]

  provisioner "local-exec" {
    command = "PGPASSWORD=${var.db_password} psql -h ${aws_db_instance.ledger_db.address} -U ${var.db_username} -d ${var.db_name} -a -f ${path.module}/init_db.sql"
  }

  # RDS yapılandırması değişirse yeniden çalıştır
  triggers = {
    db_instance_id = aws_db_instance.ledger_db.id
  }
} 