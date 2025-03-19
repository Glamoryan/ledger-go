import os
import json
import sys
import os.path

sys.path.append(os.path.dirname(os.path.realpath(__file__)))
sys.path.append('python/')

import psycopg2

def lambda_handler(event, context):
    print(f"Current directory: {os.getcwd()}")
    print(f"System path: {sys.path}")
    print(f"Directory contents: {os.listdir('.')}")
    try:
        print(f"Python directory contents: {os.listdir('python')}")
    except Exception as e:
        print(f"Cannot list python directory: {str(e)}")
    
    db_host = os.environ.get('DB_HOST')
    db_name = os.environ.get('DB_NAME')
    db_user = os.environ.get('DB_USER')
    db_password = os.environ.get('DB_PASSWORD')
    
    print(f"Connecting to database at {db_host}")
    
    sql_commands = """
-- Kullanıcılar tablosu
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    surname VARCHAR(255) NOT NULL,
    age INTEGER,
    email VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    credit DECIMAL(10, 2) DEFAULT 0.00
);

-- İşlem logları tablosu
CREATE TABLE IF NOT EXISTS transaction_logs (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    amount DECIMAL(10, 2) NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Admin kullanıcısını ekle (eğer yoksa)
INSERT INTO users (name, surname, age, email, password, credit)
SELECT 'Admin', 'User', 30, 'admin@ledger.com', '$2a$10$ZYTqWBQXY5OYzXRKJZyMXuF7HQWZoHJM1qZxUVfZn.hO.HW1Jq9Oe', 1000.00
WHERE NOT EXISTS (
    SELECT 1 FROM users WHERE email = 'admin@ledger.com'
);
"""
    
    try:
        conn = psycopg2.connect(
            host=db_host,
            database=db_name,
            user=db_user,
            password=db_password
        )
        
        conn.autocommit = True
        
        cur = conn.cursor()
        
        cur.execute(sql_commands)
        
        cur.close()
        conn.close()
        
        return {
            'statusCode': 200,
            'body': json.dumps('Veritabanı şeması ve ilk veriler başarıyla oluşturuldu!')
        }
    except Exception as e:
        print(f"Error: {str(e)}")
        return {
            'statusCode': 500,
            'body': json.dumps(f'Hata: {str(e)}')
        } 