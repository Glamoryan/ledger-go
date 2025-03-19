# Ledger API Terraform Projesi

Bu Terraform projesi, Go ile yazılmış Ledger API'sini AWS serverless mimarisinde çalıştırmak için gerekli altyapıyı oluşturur.

## Mimari Yapı

- **VPC ve Subnet**: Uygulama için izole bir ağ ortamı
- **RDS PostgreSQL**: Veritabanı sistemi
- **API Gateway**: API isteklerini karşılar ve Lambda fonksiyonuna veya SQS kuyruğuna yönlendirir
- **SQS**: Asenkron işlenmesi gereken istekleri kuyruğa alır
- **Lambda**: API isteklerini ve SQS mesajlarını işler
- **CloudWatch Logs**: Log kayıtlarını tutar

## Ön Gereksinimler

- Terraform (>= 1.0.0)
- AWS CLI yapılandırılmış ve yetkili
- Go (>= 1.18)
- PostgreSQL Client (psql) - Veritabanı şemasını oluşturmak için

## Kurulum ve Kullanım

### 1. Lambda Fonksiyonu Derleme

Lambda fonksiyonunu AWS için derlemek:

```bash
./build_lambda.sh
```

### 2. Terraform Değişkenlerini Düzenleme

`terraform.tfvars` dosyasındaki değerleri gözden geçirin:

```hcl
region          = "eu-north-1"           # AWS bölgesi
app_name        = "ledger"              # Uygulama adı
environment     = "dev"                 # Ortam (dev, prod, test)
lambda_zip_path = "../lambda/deployment.zip"  # Lambda derleme çıktısı
db_name         = "ledgerdb"            # Veritabanı adı
db_username     = "ledgeradmin"         # Veritabanı kullanıcı adı
vpc_cidr        = "10.0.0.0/16"         # VPC CIDR bloğu
```

### 3. Terraform Başlatma

```bash
terraform init
```

### 4. Terraform Plan

Yapılacak değişiklikleri görmek için:

```bash
terraform plan -var-file=terraform.tfvars -var="db_password=YOUR_DB_PASSWORD"
```

### 5. Terraform Apply

Altyapıyı oluşturmak için:

```bash
terraform apply -var-file=terraform.tfvars -var="db_password=YOUR_DB_PASSWORD"
```

### 6. Veritabanı Tabloları ve İlk Veri

Terraform, terraform/rds.tf dosyasındaki null_resource sayesinde otomatik olarak:
- users ve transaction_logs tablolarını oluşturur
- admin@ledger.com (şifre: admin123) kullanıcısını ekler

### 7. Altyapı Kaldırma

Oluşturulan tüm AWS kaynaklarını kaldırmak için:

```bash
terraform destroy -var-file=terraform.tfvars -var="db_password=YOUR_DB_PASSWORD"
```

## API Kullanımı

Terraform çıktıları API URL ve API Key gibi bilgileri verecektir. API'yi kullanmak için örnek curl komutları:

### Kullanıcı Ekleme

```bash
curl --location 'https://0pjr6fzik8.execute-api.eu-north-1.amazonaws.com/dev/register' \
--header 'Content-Type: application/json' \
--header 'x-api-key: g1nJ2Nw9KB6LD6vlwqcPv2Dti555DjJK7LRNx0o6' \
--data-raw '{
  "name": "Test",
  "surname": "User",
  "age": 25,
  "email": "test@example.com",
  "password": "password123"
}'
```

### Oturum Açma

```bash
curl --location 'https://0pjr6fzik8.execute-api.eu-north-1.amazonaws.com/dev/login' \
--header 'Content-Type: application/json' \
--header 'x-api-key: g1nJ2Nw9KB6LD6vlwqcPv2Dti555DjJK7LRNx0o6' \
--data-raw '{"email": "admin@ledger.com", "password": "admin123"}'
```

### Kredi Bakiyesi Görüntüleme

```bash
curl --location 'https://0pjr6fzik8.execute-api.eu-north-1.amazonaws.com/dev/users/get-credit?id=1' \
--header 'Content-Type: application/json' \
--header 'x-api-key: g1nJ2Nw9KB6LD6vlwqcPv2Dti555DjJK7LRNx0o6'
```

### Kredi Gönderme

```bash
curl --location 'https://0pjr6fzik8.execute-api.eu-north-1.amazonaws.com/dev/users/send-credit' \
--header 'Content-Type: application/json' \
--header 'x-api-key: g1nJ2Nw9KB6LD6vlwqcPv2Dti555DjJK7LRNx0o6' \
--data '{
    "sender_id": 1,
    "receiver_id": 2,
    "amount": 50,
    "description": "Test kredi gönderimi"
}'
```