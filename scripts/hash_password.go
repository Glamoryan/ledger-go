package main

import (
	"fmt"
	"log"
	"os"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Kullanım: go run scripts/hash_password.go [şifre]")
		os.Exit(1)
	}

	password := os.Args[1]
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Şifre hashleme hatası: %v", err)
	}

	fmt.Printf("Hash: %s\n", string(hashedPassword))
}
