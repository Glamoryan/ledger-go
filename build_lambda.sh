set -e

LAMBDA_DIR="lambda"
OUTPUT_DIR="lambda"
OUTPUT_FILE="deployment.zip"

echo "Lambda fonksiyonu derleniyor..."

export GOOS=linux
export GOARCH=amd64
export CGO_ENABLED=0

cd $LAMBDA_DIR

echo "Go uygulaması derleniyor..."
go build -tags lambda.norpc -o bootstrap main.go

echo "Derlenen dosya arşivleniyor..."
zip -j "$OUTPUT_FILE" bootstrap

echo "Geçici dosyalar temizleniyor..."
rm bootstrap

echo "Lambda fonksiyonu başarıyla derlendi: $LAMBDA_DIR/$OUTPUT_FILE" 