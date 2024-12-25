# Ledger API

## Authentication

The API uses JWT (JSON Web Token) for authentication. To access protected endpoints, you need to:
1. Register a user
2. Login to get a token
3. Include the token in the Authorization header for subsequent requests

#### Register a New User
```bash
curl -X POST "http://localhost:8080/users/add-user" -H "Content-Type: application/json" -d '{"name": "John", "surname": "Doe", "age": 30, "email": "john@example.com", "password": "password123"}'
```

#### Login
```bash
curl -X POST "http://localhost:8080/login" -H "Content-Type: application/json" -d '{"email": "john@example.com", "password": "password123"}'
```

#### View Your Credit Balance
```bash
curl -X GET "http://localhost:8080/users/get-credit?id=YOUR_ID" -H "Authorization: Bearer YOUR_TOKEN"
```

#### Send Credit to Another User
```bash
curl -X POST "http://localhost:8080/users/send-credit?senderId=YOUR_ID&receiverId=RECEIVER_ID&amount=50" -H "Authorization: Bearer YOUR_TOKEN"
```

#### View Transaction History
```bash
curl -X GET "http://localhost:8080/users/transaction-logs/sender?senderId=YOUR_ID&date=2024-03-20" -H "Authorization: Bearer YOUR_TOKEN"
```

#### Get User Details
```bash
curl -X GET "http://localhost:8080/users/get-user?id=YOUR_ID" -H "Authorization: Bearer YOUR_TOKEN"
```

### Admin Only Endpoints

#### Add Credit to Any User
```bash
curl -X POST "http://localhost:8080/users/add-credit?id=USER_ID&amount=100" -H "Authorization: Bearer ADMIN_TOKEN"
```

#### View All Users' Credits
```bash
curl -X GET "http://localhost:8080/users/credits" -H "Authorization: Bearer ADMIN_TOKEN"
```

#### List All Users
```bash
curl -X GET "http://localhost:8080/users" -H "Authorization: Bearer ADMIN_TOKEN"
```

### Batch Operations (Admin Only)

#### Get Multiple User Credits
```bash
curl -X POST "http://localhost:8080/users/batch/credits" \
     -H "Authorization: Bearer ADMIN_TOKEN" \
     -H "Content-Type: application/json" \
     -d '[1, 2, 3]'  # Array of user IDs
```

#### Batch Credit Update
```bash
curl -X POST "http://localhost:8080/users/batch/update-credits" \
     -H "Authorization: Bearer ADMIN_TOKEN" \
     -H "Content-Type: application/json" \
     -d '{
         "transactions": [
             {"user_id": 1, "amount": 100.50},
             {"user_id": 2, "amount": 200.75},
             {"user_id": 3, "amount": 50.25}
         ]
     }'
```

Response format for batch credit update:
```json
[
    {
        "success": true,
        "user_id": 1,
        "amount": 100.50,
        "error": ""
    },
    {
        "success": true,
        "user_id": 2,
        "amount": 200.75,
        "error": ""
    },
    {
        "success": false,
        "user_id": 3,
        "amount": 50.25,
        "error": "User not found"
    }
]
```

## Database Schema

### Users Table
```sql
CREATE TABLE users (
    ID SERIAL PRIMARY KEY,
    Name VARCHAR(255) NOT NULL,
    Surname VARCHAR(255) NOT NULL,
    Age INT NOT NULL,
    Email VARCHAR(255) UNIQUE NOT NULL,
    Password_Hash VARCHAR(255) NOT NULL,
    Role VARCHAR(10) DEFAULT 'user',
    Credit DOUBLE PRECISION DEFAULT 0
);
```

### Transaction Logs Table
```sql
CREATE TABLE transaction_logs (
    id SERIAL PRIMARY KEY,
    sender_id INT NOT NULL,
    receiver_id INT NOT NULL,
    amount DECIMAL(10, 2) NOT NULL,
    sender_credit_before DECIMAL(10, 2) NOT NULL,
    receiver_credit_before DECIMAL(10, 2) NOT NULL,
    transaction_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (sender_id) REFERENCES users(id),
    FOREIGN KEY (receiver_id) REFERENCES users(id)
);
```