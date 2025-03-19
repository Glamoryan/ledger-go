package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	_ "github.com/lib/pq"

	"ledger-lambda/src/factory"
	"ledger-lambda/src/handlers"
	"ledger-lambda/src/models"
)

type User struct {
	ID           uint    `json:"id"`
	Name         string  `json:"name"`
	Surname      string  `json:"surname"`
	Age          int     `json:"age"`
	Email        string  `json:"email"`
	PasswordHash string  `json:"-"`
	Role         string  `json:"role"`
	Credit       float64 `json:"credit"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Name     string `json:"name"`
	Surname  string `json:"surname"`
	Age      int    `json:"age"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type TransactionLog struct {
	ID                   uint      `json:"id"`
	SenderID             uint      `json:"sender_id"`
	ReceiverID           uint      `json:"receiver_id"`
	Amount               float64   `json:"amount"`
	SenderCreditBefore   float64   `json:"sender_credit_before"`
	ReceiverCreditBefore float64   `json:"receiver_credit_before"`
	TransactionDate      time.Time `json:"transaction_date"`
}

type BatchTransaction struct {
	UserID uint    `json:"user_id"`
	Amount float64 `json:"amount"`
}

type BatchTransactionResult struct {
	Success bool    `json:"success"`
	UserID  uint    `json:"user_id"`
	Amount  float64 `json:"amount"`
	Error   string  `json:"error"`
}

type Config struct {
	DB      *sql.DB
	Factory factory.Factory
}

var appConfig Config

func init() {
	var db *sql.DB
	var err error
	if dbHost := os.Getenv("DB_HOST"); dbHost != "" && dbHost != "skip-db-connection" {
		db, err = connectDB()
		if err != nil {
			log.Printf("Veritabanı bağlantısı oluşturulamadı: %v", err)
		} else {
			log.Printf("Veritabanı bağlantısı başarılı: %s", dbHost)
		}
	} else {
		log.Printf("DB_HOST ortam değişkeni bulunamadı veya skip-db-connection olarak ayarlandı")
	}

	var appFactory factory.Factory
	if db != nil {
		appFactory = factory.NewFactory(db)
		log.Printf("Factory başarıyla oluşturuldu")
	} else {
		log.Printf("Factory oluşturulamadı: veritabanı bağlantısı yok")
	}

	appConfig = Config{
		DB:      db,
		Factory: appFactory,
	}
	log.Printf("Uygulama yapılandırması tamamlandı")
}

func connectDB() (*sql.DB, error) {
	dbHost := os.Getenv("DB_HOST")
	dbPort := "5432" // PostgreSQL default port
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=require",
		dbHost, dbPort, dbUser, dbPass, dbName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("veritabanı bağlantısı oluşturulurken hata: %v", err)
	}

	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("veritabanına ping yapılırken hata: %v", err)
	}

	return db, nil
}

func handleRequest(ctx context.Context, event interface{}) (interface{}, error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("HATA - PANIC YAKALANDI: %v", r)
		}
	}()

	eventJSON, _ := json.Marshal(event)
	log.Printf("Gelen olay: %s", string(eventJSON))

	if records, ok := event.(map[string]interface{})["Records"].([]interface{}); ok {
		if len(records) > 0 {
			var sqsEvent events.SQSEvent
			if err := json.Unmarshal(eventJSON, &sqsEvent); err == nil && len(sqsEvent.Records) > 0 {
				log.Printf("SQS olay algılandı, %d kayıt ile yeniden yapılandırıldı", len(sqsEvent.Records))
				if err := handleSQSEvent(ctx, sqsEvent); err != nil {
					log.Printf("SQS olay işleme hatası: %v", err)
					return map[string]string{"status": "error", "message": err.Error()}, nil
				}
				return map[string]string{"status": "success"}, nil
			}
		}
	}

	var apiGwRequest events.APIGatewayProxyRequest
	if err := json.Unmarshal(eventJSON, &apiGwRequest); err == nil && apiGwRequest.Path != "" {
		log.Printf("API Gateway-ProxyRequest ayrıştırıldı: %s %s", apiGwRequest.HTTPMethod, apiGwRequest.Path)
		return handleAPIRequest(ctx, apiGwRequest)
	}

	switch e := event.(type) {
	case events.APIGatewayProxyRequest:
		log.Printf("API Gateway isteği alındı: %s %s", e.HTTPMethod, e.Path)
		return handleAPIRequest(ctx, e)
	case events.SQSEvent:
		log.Printf("SQS olayı alındı, %d kayıt", len(e.Records))
		if err := handleSQSEvent(ctx, e); err != nil {
			log.Printf("SQS olay işleme hatası: %v", err)
			return map[string]string{"status": "error", "message": err.Error()}, nil
		}
		return map[string]string{"status": "success"}, nil
	default:
		log.Printf("Desteklenmeyen olay türü: %T", event)
		var jsonRequest map[string]interface{}
		if err := json.Unmarshal(eventJSON, &jsonRequest); err == nil {
			if path, ok := jsonRequest["path"].(string); ok {
				if httpMethod, ok := jsonRequest["httpMethod"].(string); ok {
					log.Printf("JSON'dan algılanan API isteği: %s %s", httpMethod, path)

					proxyReq := events.APIGatewayProxyRequest{
						Path:       path,
						HTTPMethod: httpMethod,
					}

					if headers, ok := jsonRequest["headers"].(map[string]interface{}); ok {
						proxyReq.Headers = make(map[string]string)
						for k, v := range headers {
							if strVal, ok := v.(string); ok {
								proxyReq.Headers[k] = strVal
							}
						}
					}

					if qs, ok := jsonRequest["queryStringParameters"].(map[string]interface{}); ok {
						proxyReq.QueryStringParameters = make(map[string]string)
						for k, v := range qs {
							if strVal, ok := v.(string); ok {
								proxyReq.QueryStringParameters[k] = strVal
							}
						}
					}

					if body, ok := jsonRequest["body"].(string); ok {
						proxyReq.Body = body
					}

					return handleAPIRequest(ctx, proxyReq)
				}
			}
		}

		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       "Desteklenmeyen olay türü",
		}, nil
	}
}

func handleAPIRequest(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if appConfig.Factory == nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       "Factory başlatılamadı",
		}, nil
	}

	userHandler := appConfig.Factory.NewUserHandler()

	path := req.Path
	if path == "" {
		path = "/"
	}

	pathParts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	var endpoint string
	if len(pathParts) > 0 && pathParts[0] != "" {
		endpoint = pathParts[0]
	}

	switch endpoint {
	case "": // kök endpoint
		return handleRootEndpoint(ctx, req)
	case "users":
		return handleUsersEndpoint(ctx, req, pathParts[1:], userHandler)
	case "login":
		return handleLoginEndpoint(ctx, req, userHandler)
	case "register":
		return handleRegisterEndpoint(ctx, req, userHandler)
	default:
		return events.APIGatewayProxyResponse{
			StatusCode: 404,
			Body:       fmt.Sprintf("Endpoint bulunamadı: %s", path),
		}, nil
	}
}

func handleRootEndpoint(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: `{"status":"ok","message":"Ledger API çalışıyor"}`,
	}, nil
}

func handleUsersEndpoint(ctx context.Context, req events.APIGatewayProxyRequest, pathParts []string, handler *handlers.UserHandler) (events.APIGatewayProxyResponse, error) {
	mockResponseWriter := newMockResponseWriter()
	mockRequest, err := createMockRequest(req)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       fmt.Sprintf("İstek oluşturma hatası: %v", err),
		}, nil
	}

	if len(pathParts) == 0 {
		if req.HTTPMethod == "GET" {
			handler.GetAllUsers(mockResponseWriter, mockRequest)
		} else {
			return events.APIGatewayProxyResponse{
				StatusCode: 400,
				Body:       "Geçersiz istek",
			}, nil
		}
	} else {
		subPath := pathParts[0]
		switch subPath {
		case "add-user":
			if req.HTTPMethod == "POST" {
				handler.CreateUser(mockResponseWriter, mockRequest)
			}
		case "get-credit":
			if req.HTTPMethod == "GET" {
				handler.GetCredit(mockResponseWriter, mockRequest)
			}
		case "send-credit":
			if req.HTTPMethod == "POST" {
				handler.SendCredit(mockResponseWriter, mockRequest)
			}
		case "transaction-logs":
			if len(pathParts) > 1 && pathParts[1] == "sender" && req.HTTPMethod == "GET" {
				handler.GetTransactionLogsBySenderAndDate(mockResponseWriter, mockRequest)
			}
		case "get-user":
			if req.HTTPMethod == "GET" {
				handler.GetUserByID(mockResponseWriter, mockRequest)
			}
		case "add-credit":
			if req.HTTPMethod == "POST" {
				handler.AddCredit(mockResponseWriter, mockRequest)
			}
		case "credits":
			if req.HTTPMethod == "GET" {
				handler.GetAllCredits(mockResponseWriter, mockRequest)
			}
		case "batch":
			if len(pathParts) > 1 {
				switch pathParts[1] {
				case "credits":
					if req.HTTPMethod == "POST" {
						handler.GetMultipleUserCredits(mockResponseWriter, mockRequest)
					}
				case "update-credits":
					if req.HTTPMethod == "POST" {
						handler.ProcessBatchCreditUpdate(mockResponseWriter, mockRequest)
					}
				}
			}
		default:
			return events.APIGatewayProxyResponse{
				StatusCode: 404,
				Body:       fmt.Sprintf("Endpoint bulunamadı: %s", req.Path),
			}, nil
		}
	}

	return events.APIGatewayProxyResponse{
		StatusCode: mockResponseWriter.statusCode,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(mockResponseWriter.body),
	}, nil
}

func handleLoginEndpoint(ctx context.Context, req events.APIGatewayProxyRequest, handler *handlers.UserHandler) (events.APIGatewayProxyResponse, error) {
	if req.HTTPMethod != "POST" {
		return events.APIGatewayProxyResponse{
			StatusCode: 405,
			Body:       "Metod izni yok",
		}, nil
	}

	mockResponseWriter := newMockResponseWriter()
	mockRequest, err := createMockRequest(req)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       fmt.Sprintf("İstek oluşturma hatası: %v", err),
		}, nil
	}

	handler.Login(mockResponseWriter, mockRequest)

	return events.APIGatewayProxyResponse{
		StatusCode: mockResponseWriter.statusCode,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(mockResponseWriter.body),
	}, nil
}

func handleRegisterEndpoint(ctx context.Context, req events.APIGatewayProxyRequest, handler *handlers.UserHandler) (events.APIGatewayProxyResponse, error) {
	if req.HTTPMethod != "POST" {
		return events.APIGatewayProxyResponse{
			StatusCode: 405,
			Body:       "Metod izni yok",
		}, nil
	}

	mockResponseWriter := newMockResponseWriter()
	mockRequest, err := createMockRequest(req)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       fmt.Sprintf("İstek oluşturma hatası: %v", err),
		}, nil
	}

	handler.CreateUser(mockResponseWriter, mockRequest)

	return events.APIGatewayProxyResponse{
		StatusCode: mockResponseWriter.statusCode,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(mockResponseWriter.body),
	}, nil
}

func handleSQSEvent(ctx context.Context, event events.SQSEvent) error {
	log.Printf("SQS Event alındı. Kayıt sayısı: %d", len(event.Records))

	for i, record := range event.Records {
		log.Printf("SQS mesajı #%d işleniyor: %s, Body: %s", i+1, record.MessageId, record.Body)

		var body map[string]interface{}
		if err := json.Unmarshal([]byte(record.Body), &body); err != nil {
			log.Printf("SQS mesajı ayrıştırılamadı: %v, Raw body: %s", err, record.Body)
			continue
		}

		log.Printf("SQS mesajı ayrıştırıldı. Body içeriği: %+v", body)

		path, ok := body["path"].(string)
		if !ok {
			log.Printf("SQS mesajında path alanı bulunamadı. Body içeriği: %+v", body)
			continue
		}

		httpMethod, ok := body["httpMethod"].(string)
		if !ok {
			log.Printf("SQS mesajında httpMethod alanı bulunamadı. Body içeriği: %+v", body)
			continue
		}

		bodyData, hasBody := body["body"]
		if !hasBody {
			log.Printf("SQS mesajında body alanı bulunamadı. Body içeriği: %+v", body)
			continue
		}

		log.Printf("SQS mesajı işleniyor. Path: %s, Method: %s, Body data type: %T", path, httpMethod, bodyData)

		if path == "/users/add-user" && httpMethod == "POST" && hasBody {
			log.Printf("Kullanıcı ekleme isteği işleniyor. Body data: %+v", bodyData)

			var userData models.RegisterRequest
			bodyJSON, err := json.Marshal(bodyData)
			if err != nil {
				log.Printf("Body JSON'a dönüştürülemedi: %v", err)
				continue
			}

			log.Printf("Body JSON string: %s", string(bodyJSON))

			if err := json.Unmarshal(bodyJSON, &userData); err != nil {
				log.Printf("JSON RegisterRequest'e dönüştürülemedi: %v, JSON: %s", err, string(bodyJSON))
				continue
			}

			log.Printf("RegisterRequest ayrıştırıldı: %+v", userData)

			user := &models.User{
				Name:     userData.Name,
				Surname:  userData.Surname,
				Age:      userData.Age,
				Email:    userData.Email,
				Password: userData.Password,
			}

			log.Printf("Kullanıcı modeli oluşturuldu: %+v", user)

			userService := appConfig.Factory.NewUserService()
			if err := userService.CreateUser(user); err != nil {
				log.Printf("Kullanıcı oluşturma hatası: %v", err)
				continue
			}

			log.Printf("Kullanıcı başarıyla oluşturuldu: %s (ID: %d)", user.Email, user.ID)
		} else {
			log.Printf("İşlenmeyen SQS mesajı: Path=%s, Method=%s", path, httpMethod)
		}
	}
	return nil
}

type mockResponseWriter struct {
	headers    http.Header
	body       []byte
	statusCode int
}

func newMockResponseWriter() *mockResponseWriter {
	return &mockResponseWriter{
		headers:    make(http.Header),
		statusCode: http.StatusOK,
	}
}

func (m *mockResponseWriter) Header() http.Header {
	return m.headers
}

func (m *mockResponseWriter) Write(body []byte) (int, error) {
	m.body = body
	return len(body), nil
}

func (m *mockResponseWriter) WriteHeader(statusCode int) {
	m.statusCode = statusCode
}

func createMockRequest(req events.APIGatewayProxyRequest) (*http.Request, error) {
	query := url.Values{}
	for k, v := range req.QueryStringParameters {
		query.Set(k, v)
	}

	var urlStr string
	if strings.HasPrefix(req.Path, "/") {
		urlStr = "http://localhost" + req.Path
	} else {
		urlStr = "http://localhost/" + req.Path
	}

	if len(query) > 0 {
		urlStr = urlStr + "?" + query.Encode()
	}

	var httpReq *http.Request
	var err error

	if req.Body != "" {
		httpReq, err = http.NewRequest(req.HTTPMethod, urlStr, strings.NewReader(req.Body))
	} else {
		httpReq, err = http.NewRequest(req.HTTPMethod, urlStr, nil)
	}

	if err != nil {
		return nil, err
	}

	for k, v := range req.Headers {
		httpReq.Header.Set(k, v)
	}

	return httpReq, nil
}

func main() {
	lambda.Start(handleRequest)
}
