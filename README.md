### Komutlar
```bash
# Add User
curl -X POST -H "Content-Type: application/json" -d '{"name": "John", "surname": "Doe", "age": 30}' http://localhost:8080/users/add-user

# List All Users
curl -X GET http://localhost:8080/users

# Get Specific User
curl -X GET "http://localhost:8080/users/get-user?id=1"

# Add Credit to User
curl -X POST "http://localhost:8080/users/add-credit?id=1&amount=50"

# Get All Users' Credits
curl -X GET "http://localhost:8080/users/credits"

# Query User's Credit
curl -X GET "http://localhost:8080/users/get-credit?id=1"

# Missing or Invalid `id`
curl -X POST "http://localhost:8080/users/add-credit?id=abc&amount=50"

# Missing or Invalid `amount`
curl -X POST "http://localhost:8080/users/add-credit?id=1&amount=abc"

# Send credit to a user
curl --location --request POST 'http://localhost:8080/users/send-credit?senderId=1&receiverId=4&amount=50' \
--header 'Content-Type: application/json' \
--data ''
```
### DB
```sql
-- Create table for users
CREATE TABLE users (
    ID SERIAL PRIMARY KEY,         
    Name VARCHAR(255) NOT NULL,    
    Surname VARCHAR(255) NOT NULL, 
    Age INT NOT NULL,              
    Credit DOUBLE PRECISION DEFAULT 0 
);

-- Create table for transaction logs
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


-- Add a sample user
INSERT INTO users (name, surname, age, credit)
VALUES ('John', 'Doe', 30, 50.0);
