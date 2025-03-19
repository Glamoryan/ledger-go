#!/bin/bash
set -e

echo "Lambda fonksiyonu derleniyor..."

cd ../lambda

# Go modüllerini indirin
go mod tidy

# Linux için derle (AWS Lambda ortamı)
GOOS=linux GOARCH=amd64 go build -o bootstrap main.go

# Çalıştırma izni ekleyin
chmod +x bootstrap

# Deployment paketini oluşturun
zip deployment.zip bootstrap

# Ana dizine dönün
cd ..

echo "Lambda deployment paketi başarıyla oluşturuldu: lambda/deployment.zip" 