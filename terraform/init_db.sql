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