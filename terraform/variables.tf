variable "region" {
  description = "AWS bölgesi"
  type        = string
  default     = "eu-north-1"
}

variable "app_name" {
  description = "Uygulama adı"
  type        = string
  default     = "ledger"
}

variable "environment" {
  description = "Deployment ortamı"
  type        = string
  default     = "dev"
}

variable "lambda_zip_path" {
  description = "Lambda fonksiyonu için zip dosyasının yolu"
  type        = string
  default     = "lambda/deployment.zip"
}

variable "db_name" {
  description = "Veritabanı adı"
  type        = string
  default     = "ledgerdb"
}

variable "db_username" {
  description = "Veritabanı kullanıcı adı"
  type        = string
  default     = "ledgeradmin"
}

variable "db_password" {
  description = "Veritabanı şifresi"
  type        = string
  sensitive   = true
}

variable "vpc_cidr" {
  description = "VPC için CIDR bloğu"
  type        = string
  default     = "10.0.0.0/16"
} 