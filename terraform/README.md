# Ledger API Terraform

## Ön Gereksinimler

- Terraform (>= 1.0.0)
- AWS CLI yapılandırılmış ve yetkili
- Go (>= 1.18)
- PostgreSQL Client (psql) - Veritabanı şemasını oluşturmak için

## API Kullanımı

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
