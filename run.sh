set -e

REGION="eu-north-1"
APP_NAME="ledger"
ENVIRONMENT="dev"
DB_PASSWORD="Test123456"  # Gerçek projede daha güvenli bir yöntem kullanmalısınız

echo "=== Lambda derleniyor... ==="
./build_lambda.sh

echo "=== Terraform başlatılıyor... ==="
cd terraform
terraform init

echo "=== Altyapı planı oluşturuluyor... ==="
terraform plan -var="db_password=$DB_PASSWORD"

echo "=== Altyapı oluşturuluyor... ==="
terraform apply -var="db_password=$DB_PASSWORD" -auto-approve

echo "=== Tamamlandı! ==="
terraform output

echo ""
echo "API URL: $(terraform output -raw api_url)"
echo "API_KEY: (terraform output -raw api_key)"
echo ""

echo "API kullanım örneği:"
echo "curl -X GET \"$(terraform output -raw api_url)/users/get-credit?id=1\" -H \"x-api-key: $(terraform output -raw api_key)\"" 