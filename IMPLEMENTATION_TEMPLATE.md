# BACKEND IMPLEMENTATION TEMPLATE
## Hướng Dẫn Triển Khai Khuôn Mẫu Backend System

> **Mục tiêu**: Template này cung cấp code examples và implementation guide cụ thể để triển khai checklist backend system.

---

## 🏗️ **1. PROJECT SETUP TEMPLATE**

### Go Project Structure
```bash
# Khởi tạo project
mkdir my-backend-system
cd my-backend-system
go mod init my-backend-system

# Tạo cấu trúc thư mục
mkdir -p {configs,src/{bootstrap,common,core,infra,present},tests,scripts,docs}
mkdir -p src/common/{cache,configs,crypto,logger,storage,tracer,fault}
mkdir -p src/core/{domains,enums,usecases}
mkdir -p src/infra/{database,external,cache}
mkdir -p src/present/{http,consumers}
mkdir -p src/present/http/{controllers,middlewares,router}
```

### Essential Dependencies (go.mod)
```go
module my-backend-system

go 1.21

require (
    // HTTP Framework
    github.com/gin-gonic/gin v1.9.1
    
    // Dependency Injection
    go.uber.org/fx v1.20.0
    
    // Configuration
    github.com/spf13/viper v1.16.0
    
    // Database
    go.mongodb.org/mongo-driver v1.12.1
    github.com/lib/pq v1.10.9 // PostgreSQL
    
    // Cache
    github.com/redis/go-redis/v9 v9.1.0
    github.com/dgraph-io/ristretto v0.1.1
    
    // Logging
    github.com/rs/zerolog v1.30.0
    
    // Tracing
    go.opentelemetry.io/otel v1.16.0
    go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin v0.42.0
    
    // Crypto
    golang.org/x/crypto v0.12.0
    github.com/golang-jwt/jwt/v5 v5.0.0
    
    // Storage
    github.com/aws/aws-sdk-go v1.44.327
    github.com/minio/minio-go/v7 v7.0.61
    
    // Validation
    github.com/go-playground/validator/v10 v10.15.1
    
    // HTTP Client
    github.com/imroc/req/v3 v3.41.0
    
    // Testing
    github.com/stretchr/testify v1.8.4
    github.com/testcontainers/testcontainers-go v0.23.0
)
```

---

## 🔧 **2. CONFIGURATION TEMPLATE**

### configs/config.yaml
```yaml
# Application Settings
app:
  name: "my-backend-system"
  version: "1.0.0"
  environment: "development"
  debug: true

# Server Configuration
server:
  host: "0.0.0.0"
  port: 8080
  read_timeout: "30s"
  write_timeout: "30s"
  idle_timeout: "120s"
  shutdown_timeout: "10s"

# Database Configuration
database:
  type: "postgres" # postgres, mysql, mongodb
  host: "localhost"
  port: 5432
  name: "myapp_db"
  username: "postgres"
  password: "password"
  ssl_mode: "disable"
  max_open_conns: 25
  max_idle_conns: 5
  conn_max_lifetime: "1h"

# Cache Configuration
cache:
  redis:
    host: "localhost"
    port: 6379
    password: ""
    db: 0
    pool_size: 10
    min_idle_conns: 5
  memory:
    max_cost: 1073741824 # 1GB
    num_counters: 1000000

# Security Configuration
security:
  jwt:
    access_secret: "your-access-secret-key"
    refresh_secret: "your-refresh-secret-key"
    access_expire: "15m"
    refresh_expire: "7d"
  api_keys:
    admin: "admin-api-key-here"
    service: "service-api-key-here"
  encryption:
    aes_key: "32-char-aes-key-here-12345678901"
    rsa_private_key_path: "configs/rsa_private.pem"
    rsa_public_key_path: "configs/rsa_public.pem"

# External Services
external:
  email:
    provider: "sendgrid"
    api_key: "your-sendgrid-api-key"
    from_email: "noreply@yourapp.com"
  sms:
    provider: "twilio"
    account_sid: "your-twilio-sid"
    auth_token: "your-twilio-token"
  storage:
    provider: "s3"
    bucket: "your-bucket-name"
    region: "us-west-2"
    access_key: "your-access-key"
    secret_key: "your-secret-key"

# Logging Configuration
logging:
  level: "info" # debug, info, warn, error
  format: "json" # json, text
  output: "stdout" # stdout, stderr, file
  file_path: "logs/app.log"
  max_size: 100 # MB
  max_backups: 5
  max_age: 30 # days

# Monitoring Configuration
monitoring:
  metrics:
    enabled: true
    path: "/metrics"
  tracing:
    enabled: true
    jaeger_endpoint: "http://localhost:14268/api/traces"
    sample_rate: 0.1
  health_check:
    path: "/health"
    interval: "30s"

# Feature Flags
features:
  new_user_flow: true
  advanced_analytics: false
  beta_features: false
```

### Configuration Loader (src/common/configs/config.go)
```go
package configs

import (
    "fmt"
    "time"
    
    "github.com/spf13/viper"
)

type Config struct {
    App        AppConfig        `mapstructure:"app"`
    Server     ServerConfig     `mapstructure:"server"`
    Database   DatabaseConfig   `mapstructure:"database"`
    Cache      CacheConfig      `mapstructure:"cache"`
    Security   SecurityConfig   `mapstructure:"security"`
    External   ExternalConfig   `mapstructure:"external"`
    Logging    LoggingConfig    `mapstructure:"logging"`
    Monitoring MonitoringConfig `mapstructure:"monitoring"`
    Features   FeatureConfig    `mapstructure:"features"`
}

type AppConfig struct {
    Name        string `mapstructure:"name"`
    Version     string `mapstructure:"version"`
    Environment string `mapstructure:"environment"`
    Debug       bool   `mapstructure:"debug"`
}

type ServerConfig struct {
    Host            string        `mapstructure:"host"`
    Port            int           `mapstructure:"port"`
    ReadTimeout     time.Duration `mapstructure:"read_timeout"`
    WriteTimeout    time.Duration `mapstructure:"write_timeout"`
    IdleTimeout     time.Duration `mapstructure:"idle_timeout"`
    ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}

type DatabaseConfig struct {
    Type            string        `mapstructure:"type"`
    Host            string        `mapstructure:"host"`
    Port            int           `mapstructure:"port"`
    Name            string        `mapstructure:"name"`
    Username        string        `mapstructure:"username"`
    Password        string        `mapstructure:"password"`
    SSLMode         string        `mapstructure:"ssl_mode"`
    MaxOpenConns    int           `mapstructure:"max_open_conns"`
    MaxIdleConns    int           `mapstructure:"max_idle_conns"`
    ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

// ... other config structs

func LoadConfig(path string) (*Config, error) {
    viper.SetConfigFile(path)
    viper.AutomaticEnv()
    
    if err := viper.ReadInConfig(); err != nil {
        return nil, fmt.Errorf("failed to read config: %w", err)
    }
    
    var config Config
    if err := viper.Unmarshal(&config); err != nil {
        return nil, fmt.Errorf("failed to unmarshal config: %w", err)
    }
    
    return &config, nil
}

func (c *Config) IsDevelopment() bool {
    return c.App.Environment == "development"
}

func (c *Config) IsProduction() bool {
    return c.App.Environment == "production"
}
```

---

## 📊 **3. LOGGING SYSTEM TEMPLATE**

### Logger Interface (src/common/logger/logger.go)
```go
package logger

import (
    "context"
    "io"
    "os"
    "time"
    
    "github.com/rs/zerolog"
)

type Logger interface {
    Debug(ctx context.Context, msg string, fields ...Field)
    Info(ctx context.Context, msg string, fields ...Field)
    Warn(ctx context.Context, msg string, fields ...Field)
    Error(ctx context.Context, err error, msg string, fields ...Field)
    Fatal(ctx context.Context, err error, msg string, fields ...Field)
    
    With(fields ...Field) Logger
    WithContext(ctx context.Context) Logger
}

type Field struct {
    Key   string
    Value interface{}
}

func String(key, value string) Field {
    return Field{Key: key, Value: value}
}

func Int(key string, value int) Field {
    return Field{Key: key, Value: value}
}

func Duration(key string, value time.Duration) Field {
    return Field{Key: key, Value: value}
}

func Error(err error) Field {
    return Field{Key: "error", Value: err}
}

type zerologLogger struct {
    logger zerolog.Logger
}

func NewLogger(output io.Writer, level string) Logger {
    logLevel, err := zerolog.ParseLevel(level)
    if err != nil {
        logLevel = zerolog.InfoLevel
    }
    
    zerolog.TimeFieldFormat = time.RFC3339
    logger := zerolog.New(output).
        Level(logLevel).
        With().
        Timestamp().
        Caller().
        Logger()
    
    return &zerologLogger{logger: logger}
}

func NewConsoleLogger(level string) Logger {
    output := zerolog.ConsoleWriter{
        Out:        os.Stdout,
        TimeFormat: time.RFC3339,
    }
    return NewLogger(output, level)
}

func (l *zerologLogger) Debug(ctx context.Context, msg string, fields ...Field) {
    l.log(l.logger.Debug(), ctx, msg, fields...)
}

func (l *zerologLogger) Info(ctx context.Context, msg string, fields ...Field) {
    l.log(l.logger.Info(), ctx, msg, fields...)
}

func (l *zerologLogger) Warn(ctx context.Context, msg string, fields ...Field) {
    l.log(l.logger.Warn(), ctx, msg, fields...)
}

func (l *zerologLogger) Error(ctx context.Context, err error, msg string, fields ...Field) {
    event := l.logger.Error()
    if err != nil {
        event = event.Err(err)
    }
    l.log(event, ctx, msg, fields...)
}

func (l *zerologLogger) Fatal(ctx context.Context, err error, msg string, fields ...Field) {
    event := l.logger.Fatal()
    if err != nil {
        event = event.Err(err)
    }
    l.log(event, ctx, msg, fields...)
}

func (l *zerologLogger) log(event *zerolog.Event, ctx context.Context, msg string, fields ...Field) {
    // Add trace information from context
    if traceID := getTraceID(ctx); traceID != "" {
        event = event.Str("trace_id", traceID)
    }
    
    if userID := getUserID(ctx); userID != "" {
        event = event.Str("user_id", userID)
    }
    
    // Add custom fields
    for _, field := range fields {
        event = event.Interface(field.Key, field.Value)
    }
    
    event.Msg(msg)
}

func (l *zerologLogger) With(fields ...Field) Logger {
    logger := l.logger
    for _, field := range fields {
        logger = logger.With().Interface(field.Key, field.Value).Logger()
    }
    return &zerologLogger{logger: logger}
}

func (l *zerologLogger) WithContext(ctx context.Context) Logger {
    logger := l.logger
    
    if traceID := getTraceID(ctx); traceID != "" {
        logger = logger.With().Str("trace_id", traceID).Logger()
    }
    
    if userID := getUserID(ctx); userID != "" {
        logger = logger.With().Str("user_id", userID).Logger()
    }
    
    return &zerologLogger{logger: logger}
}

// Helper functions to extract context values
func getTraceID(ctx context.Context) string {
    if traceID, ok := ctx.Value("trace_id").(string); ok {
        return traceID
    }
    return ""
}

func getUserID(ctx context.Context) string {
    if userID, ok := ctx.Value("user_id").(string); ok {
        return userID
    }
    return ""
}
```

### Global Logger (src/common/logger/global.go)
```go
package logger

import (
    "context"
    "sync"
)

var (
    globalLogger Logger
    once         sync.Once
)

func InitGlobalLogger(logger Logger) {
    once.Do(func() {
        globalLogger = logger
    })
}

func Debug(ctx context.Context, msg string, fields ...Field) {
    globalLogger.Debug(ctx, msg, fields...)
}

func Info(ctx context.Context, msg string, fields ...Field) {
    globalLogger.Info(ctx, msg, fields...)
}

func Warn(ctx context.Context, msg string, fields ...Field) {
    globalLogger.Warn(ctx, msg, fields...)
}

func ErrorWithErr(ctx context.Context, err error, msg string, fields ...Field) {
    globalLogger.Error(ctx, err, msg, fields...)
}

func Fatal(ctx context.Context, err error, msg string, fields ...Field) {
    globalLogger.Fatal(ctx, err, msg, fields...)
}
```

---

## 🛡️ **4. AUTHENTICATION SYSTEM TEMPLATE**

### JWT Service (src/common/crypto/jwt/jwt.go)
```go
package jwt

import (
    "fmt"
    "time"
    
    "github.com/golang-jwt/jwt/v5"
)

type Service struct {
    accessSecret  string
    refreshSecret string
    accessExpire  time.Duration
    refreshExpire time.Duration
}

type Claims struct {
    UserID   string `json:"user_id"`
    Email    string `json:"email"`
    Role     string `json:"role"`
    IssuedAt int64  `json:"iat"`
    jwt.RegisteredClaims
}

type TokenPair struct {
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
    ExpiresIn    int64  `json:"expires_in"`
}

func NewService(accessSecret, refreshSecret string, accessExpire, refreshExpire time.Duration) *Service {
    return &Service{
        accessSecret:  accessSecret,
        refreshSecret: refreshSecret,
        accessExpire:  accessExpire,
        refreshExpire: refreshExpire,
    }
}

func (s *Service) GenerateTokenPair(userID, email, role string) (*TokenPair, error) {
    // Generate access token
    accessClaims := &Claims{
        UserID:   userID,
        Email:    email,
        Role:     role,
        IssuedAt: time.Now().Unix(),
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.accessExpire)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            NotBefore: jwt.NewNumericDate(time.Now()),
        },
    }
    
    accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
    accessTokenString, err := accessToken.SignedString([]byte(s.accessSecret))
    if err != nil {
        return nil, fmt.Errorf("failed to sign access token: %w", err)
    }
    
    // Generate refresh token
    refreshClaims := &Claims{
        UserID:   userID,
        Email:    email,
        Role:     role,
        IssuedAt: time.Now().Unix(),
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.refreshExpire)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            NotBefore: jwt.NewNumericDate(time.Now()),
        },
    }
    
    refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
    refreshTokenString, err := refreshToken.SignedString([]byte(s.refreshSecret))
    if err != nil {
        return nil, fmt.Errorf("failed to sign refresh token: %w", err)
    }
    
    return &TokenPair{
        AccessToken:  accessTokenString,
        RefreshToken: refreshTokenString,
        ExpiresIn:    int64(s.accessExpire.Seconds()),
    }, nil
}

func (s *Service) ValidateAccessToken(tokenString string) (*Claims, error) {
    return s.validateToken(tokenString, s.accessSecret)
}

func (s *Service) ValidateRefreshToken(tokenString string) (*Claims, error) {
    return s.validateToken(tokenString, s.refreshSecret)
}

func (s *Service) validateToken(tokenString, secret string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return []byte(secret), nil
    })
    
    if err != nil {
        return nil, fmt.Errorf("failed to parse token: %w", err)
    }
    
    claims, ok := token.Claims.(*Claims)
    if !ok || !token.Valid {
        return nil, fmt.Errorf("invalid token")
    }
    
    return claims, nil
}

func (s *Service) RefreshToken(refreshTokenString string) (*TokenPair, error) {
    claims, err := s.ValidateRefreshToken(refreshTokenString)
    if err != nil {
        return nil, fmt.Errorf("invalid refresh token: %w", err)
    }
    
    return s.GenerateTokenPair(claims.UserID, claims.Email, claims.Role)
}
```

### Auth Middleware (src/present/http/middlewares/auth.go)
```go
package middlewares

import (
    "context"
    "net/http"
    "strings"
    
    "github.com/gin-gonic/gin"
    
    "my-backend-system/src/common/crypto/jwt"
    "my-backend-system/src/common/logger"
)

type AuthMiddleware struct {
    jwtService *jwt.Service
    logger     logger.Logger
}

func NewAuthMiddleware(jwtService *jwt.Service, logger logger.Logger) *AuthMiddleware {
    return &AuthMiddleware{
        jwtService: jwtService,
        logger:     logger,
    }
}

func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := m.extractToken(c)
        if token == "" {
            m.logger.Warn(c.Request.Context(), "Missing authorization token")
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization token required"})
            c.Abort()
            return
        }
        
        claims, err := m.jwtService.ValidateAccessToken(token)
        if err != nil {
            m.logger.Error(c.Request.Context(), err, "Invalid authorization token")
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
            c.Abort()
            return
        }
        
        // Add user info to context
        ctx := context.WithValue(c.Request.Context(), "user_id", claims.UserID)
        ctx = context.WithValue(ctx, "user_email", claims.Email)
        ctx = context.WithValue(ctx, "user_role", claims.Role)
        c.Request = c.Request.WithContext(ctx)
        
        // Add user info to gin context
        c.Set("user_id", claims.UserID)
        c.Set("user_email", claims.Email)
        c.Set("user_role", claims.Role)
        
        c.Next()
    }
}

func (m *AuthMiddleware) RequireRole(roles ...string) gin.HandlerFunc {
    return func(c *gin.Context) {
        userRole, exists := c.Get("user_role")
        if !exists {
            c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
            c.Abort()
            return
        }
        
        roleStr := userRole.(string)
        for _, role := range roles {
            if roleStr == role {
                c.Next()
                return
            }
        }
        
        m.logger.Warn(c.Request.Context(), "Insufficient permissions", 
            logger.String("required_roles", strings.Join(roles, ",")),
            logger.String("user_role", roleStr))
        
        c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
        c.Abort()
    }
}

func (m *AuthMiddleware) extractToken(c *gin.Context) string {
    // Check Authorization header
    authHeader := c.GetHeader("Authorization")
    if authHeader != "" {
        parts := strings.SplitN(authHeader, " ", 2)
        if len(parts) == 2 && parts[0] == "Bearer" {
            return parts[1]
        }
    }
    
    // Check query parameter
    return c.Query("token")
}
```

---

## 💾 **5. DATABASE LAYER TEMPLATE**

### Database Interface (src/infra/database/database.go)
```go
package database

import (
    "context"
    "database/sql"
)

type DB interface {
    // Connection management
    Ping(ctx context.Context) error
    Close() error
    Stats() sql.DBStats
    
    // Query operations
    Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
    QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row
    Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
    
    // Transaction operations
    Begin(ctx context.Context) (Tx, error)
    BeginTx(ctx context.Context, opts *sql.TxOptions) (Tx, error)
}

type Tx interface {
    Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
    QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row
    Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
    Commit() error
    Rollback() error
}

type Repository interface {
    // Health check
    Health(ctx context.Context) error
    
    // Migration
    Migrate(ctx context.Context) error
}
```

### PostgreSQL Implementation (src/infra/database/postgres.go)
```go
package database

import (
    "context"
    "database/sql"
    "fmt"
    "time"
    
    _ "github.com/lib/pq"
    
    "my-backend-system/src/common/configs"
    "my-backend-system/src/common/logger"
)

type PostgresDB struct {
    db     *sql.DB
    config *configs.DatabaseConfig
    logger logger.Logger
}

type PostgresTx struct {
    tx *sql.Tx
}

func NewPostgresDB(config *configs.DatabaseConfig, logger logger.Logger) (*PostgresDB, error) {
    dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
        config.Host, config.Port, config.Username, config.Password, config.Name, config.SSLMode)
    
    db, err := sql.Open("postgres", dsn)
    if err != nil {
        return nil, fmt.Errorf("failed to open database: %w", err)
    }
    
    // Configure connection pool
    db.SetMaxOpenConns(config.MaxOpenConns)
    db.SetMaxIdleConns(config.MaxIdleConns)
    db.SetConnMaxLifetime(config.ConnMaxLifetime)
    
    // Test connection
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    if err := db.PingContext(ctx); err != nil {
        return nil, fmt.Errorf("failed to ping database: %w", err)
    }
    
    logger.Info(context.Background(), "Connected to PostgreSQL database",
        logger.String("host", config.Host),
        logger.Int("port", config.Port),
        logger.String("database", config.Name))
    
    return &PostgresDB{
        db:     db,
        config: config,
        logger: logger,
    }, nil
}

func (p *PostgresDB) Ping(ctx context.Context) error {
    return p.db.PingContext(ctx)
}

func (p *PostgresDB) Close() error {
    return p.db.Close()
}

func (p *PostgresDB) Stats() sql.DBStats {
    return p.db.Stats()
}

func (p *PostgresDB) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
    start := time.Now()
    rows, err := p.db.QueryContext(ctx, query, args...)
    duration := time.Since(start)
    
    p.logger.Debug(ctx, "Database query executed",
        logger.String("query", query),
        logger.Duration("duration", duration),
        logger.String("error", fmt.Sprintf("%v", err)))
    
    return rows, err
}

func (p *PostgresDB) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
    start := time.Now()
    row := p.db.QueryRowContext(ctx, query, args...)
    duration := time.Since(start)
    
    p.logger.Debug(ctx, "Database query row executed",
        logger.String("query", query),
        logger.Duration("duration", duration))
    
    return row
}

func (p *PostgresDB) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
    start := time.Now()
    result, err := p.db.ExecContext(ctx, query, args...)
    duration := time.Since(start)
    
    p.logger.Debug(ctx, "Database exec executed",
        logger.String("query", query),
        logger.Duration("duration", duration),
        logger.String("error", fmt.Sprintf("%v", err)))
    
    return result, err
}

func (p *PostgresDB) Begin(ctx context.Context) (Tx, error) {
    tx, err := p.db.BeginTx(ctx, nil)
    if err != nil {
        return nil, err
    }
    return &PostgresTx{tx: tx}, nil
}

func (p *PostgresDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (Tx, error) {
    tx, err := p.db.BeginTx(ctx, opts)
    if err != nil {
        return nil, err
    }
    return &PostgresTx{tx: tx}, nil
}

// Transaction implementation
func (t *PostgresTx) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
    return t.tx.QueryContext(ctx, query, args...)
}

func (t *PostgresTx) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
    return t.tx.QueryRowContext(ctx, query, args...)
}

func (t *PostgresTx) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
    return t.tx.ExecContext(ctx, query, args...)
}

func (t *PostgresTx) Commit() error {
    return t.tx.Commit()
}

func (t *PostgresTx) Rollback() error {
    return t.tx.Rollback()
}
```

---

## ⚡ **6. CACHE SYSTEM TEMPLATE**

### Cache Interface (src/common/cache/cache.go)
```go
package cache

import (
    "context"
    "time"
)

type Cache interface {
    // Basic operations
    Get(ctx context.Context, key string) (interface{}, error)
    Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
    Delete(ctx context.Context, key string) error
    Exists(ctx context.Context, key string) (bool, error)
    
    // Batch operations
    MGet(ctx context.Context, keys []string) (map[string]interface{}, error)
    MSet(ctx context.Context, items map[string]interface{}, ttl time.Duration) error
    MDelete(ctx context.Context, keys []string) error
    
    // Advanced operations
    Increment(ctx context.Context, key string, value int64) (int64, error)
    Decrement(ctx context.Context, key string, value int64) (int64, error)
    Expire(ctx context.Context, key string, ttl time.Duration) error
    TTL(ctx context.Context, key string) (time.Duration, error)
    
    // Pattern operations
    Keys(ctx context.Context, pattern string) ([]string, error)
    DeletePattern(ctx context.Context, pattern string) error
    
    // Health check
    Ping(ctx context.Context) error
    Close() error
}

type CacheStats struct {
    Hits        int64
    Misses      int64
    Keys        int64
    Memory      int64
    Connections int
}

type CacheManager interface {
    Cache
    Stats(ctx context.Context) (*CacheStats, error)
    FlushAll(ctx context.Context) error
}
```

### Redis Implementation (src/common/cache/redis.go)
```go
package cache

import (
    "context"
    "encoding/json"
    "fmt"
    "time"
    
    "github.com/redis/go-redis/v9"
    
    "my-backend-system/src/common/configs"
    "my-backend-system/src/common/logger"
)

type RedisCache struct {
    client redis.UniversalClient
    logger logger.Logger
}

func NewRedisCache(config *configs.CacheConfig, logger logger.Logger) (*RedisCache, error) {
    opts := &redis.Options{
        Addr:         fmt.Sprintf("%s:%d", config.Redis.Host, config.Redis.Port),
        Password:     config.Redis.Password,
        DB:           config.Redis.DB,
        PoolSize:     config.Redis.PoolSize,
        MinIdleConns: config.Redis.MinIdleConns,
    }
    
    client := redis.NewClient(opts)
    
    // Test connection
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    if err := client.Ping(ctx).Err(); err != nil {
        return nil, fmt.Errorf("failed to connect to Redis: %w", err)
    }
    
    logger.Info(context.Background(), "Connected to Redis cache",
        logger.String("addr", opts.Addr),
        logger.Int("db", opts.DB))
    
    return &RedisCache{
        client: client,
        logger: logger,
    }, nil
}

func (r *RedisCache) Get(ctx context.Context, key string) (interface{}, error) {
    start := time.Now()
    result, err := r.client.Get(ctx, key).Result()
    duration := time.Since(start)
    
    if err == redis.Nil {
        r.logger.Debug(ctx, "Cache miss",
            logger.String("key", key),
            logger.Duration("duration", duration))
        return nil, ErrCacheMiss
    }
    
    if err != nil {
        r.logger.Error(ctx, err, "Cache get error",
            logger.String("key", key),
            logger.Duration("duration", duration))
        return nil, err
    }
    
    r.logger.Debug(ctx, "Cache hit",
        logger.String("key", key),
        logger.Duration("duration", duration))
    
    var value interface{}
    if err := json.Unmarshal([]byte(result), &value); err != nil {
        return result, nil // Return as string if not JSON
    }
    
    return value, nil
}

func (r *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
    start := time.Now()
    
    var data string
    switch v := value.(type) {
    case string:
        data = v
    case []byte:
        data = string(v)
    default:
        jsonData, err := json.Marshal(value)
        if err != nil {
            return fmt.Errorf("failed to marshal value: %w", err)
        }
        data = string(jsonData)
    }
    
    err := r.client.Set(ctx, key, data, ttl).Err()
    duration := time.Since(start)
    
    if err != nil {
        r.logger.Error(ctx, err, "Cache set error",
            logger.String("key", key),
            logger.Duration("ttl", ttl),
            logger.Duration("duration", duration))
        return err
    }
    
    r.logger.Debug(ctx, "Cache set",
        logger.String("key", key),
        logger.Duration("ttl", ttl),
        logger.Duration("duration", duration))
    
    return nil
}

func (r *RedisCache) Delete(ctx context.Context, key string) error {
    start := time.Now()
    err := r.client.Del(ctx, key).Err()
    duration := time.Since(start)
    
    if err != nil {
        r.logger.Error(ctx, err, "Cache delete error",
            logger.String("key", key),
            logger.Duration("duration", duration))
        return err
    }
    
    r.logger.Debug(ctx, "Cache delete",
        logger.String("key", key),
        logger.Duration("duration", duration))
    
    return nil
}

func (r *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
    result, err := r.client.Exists(ctx, key).Result()
    return result > 0, err
}

func (r *RedisCache) Increment(ctx context.Context, key string, value int64) (int64, error) {
    return r.client.IncrBy(ctx, key, value).Result()
}

func (r *RedisCache) Decrement(ctx context.Context, key string, value int64) (int64, error) {
    return r.client.DecrBy(ctx, key, value).Result()
}

func (r *RedisCache) TTL(ctx context.Context, key string) (time.Duration, error) {
    return r.client.TTL(ctx, key).Result()
}

func (r *RedisCache) Ping(ctx context.Context) error {
    return r.client.Ping(ctx).Err()
}

func (r *RedisCache) Close() error {
    return r.client.Close()
}

// Custom error types
var (
    ErrCacheMiss = fmt.Errorf("cache miss")
)
```

---

## 🌐 **7. HTTP SERVER TEMPLATE**

### HTTP Server (src/present/http/server.go)
```go
package http

import (
    "context"
    "fmt"
    "net/http"
    "time"
    
    "github.com/gin-gonic/gin"
    "go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
    
    "my-backend-system/src/common/configs"
    "my-backend-system/src/common/logger"
    "my-backend-system/src/present/http/middlewares"
    "my-backend-system/src/present/http/routes"
)

type Server struct {
    config     *configs.ServerConfig
    logger     logger.Logger
    httpServer *http.Server
    router     *gin.Engine
}

func NewServer(config *configs.ServerConfig, logger logger.Logger) *Server {
    // Set gin mode
    if !config.Debug {
        gin.SetMode(gin.ReleaseMode)
    }
    
    router := gin.New()
    
    return &Server{
        config: config,
        logger: logger,
        router: router,
    }
}

func (s *Server) SetupMiddlewares() {
    // Recovery middleware
    s.router.Use(gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
        s.logger.Error(c.Request.Context(), fmt.Errorf("panic recovered: %v", recovered), "HTTP panic")
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
    }))
    
    // Request logging middleware
    s.router.Use(middlewares.RequestLogger(s.logger))
    
    // CORS middleware
    s.router.Use(middlewares.CORS())
    
    // OpenTelemetry middleware
    s.router.Use(otelgin.Middleware("my-backend-system"))
    
    // Request ID middleware
    s.router.Use(middlewares.RequestID())
    
    // Rate limiting middleware
    s.router.Use(middlewares.RateLimit())
    
    // Security headers middleware
    s.router.Use(middlewares.SecurityHeaders())
}

func (s *Server) SetupRoutes(routeGroups ...routes.RouteGroup) {
    // Health check endpoint
    s.router.GET("/health", s.healthCheck)
    s.router.GET("/ready", s.readinessCheck)
    
    // Metrics endpoint
    s.router.GET("/metrics", s.metricsHandler)
    
    // Setup API routes
    api := s.router.Group("/api")
    for _, group := range routeGroups {
        group.Setup(api)
    }
}

func (s *Server) Start(ctx context.Context) error {
    addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
    
    s.httpServer = &http.Server{
        Addr:         addr,
        Handler:      s.router,
        ReadTimeout:  s.config.ReadTimeout,
        WriteTimeout: s.config.WriteTimeout,
        IdleTimeout:  s.config.IdleTimeout,
    }
    
    s.logger.Info(ctx, "Starting HTTP server",
        logger.String("addr", addr),
        logger.Duration("read_timeout", s.config.ReadTimeout),
        logger.Duration("write_timeout", s.config.WriteTimeout))
    
    if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        return fmt.Errorf("failed to start HTTP server: %w", err)
    }
    
    return nil
}

func (s *Server) Stop(ctx context.Context) error {
    s.logger.Info(ctx, "Stopping HTTP server")
    
    shutdownCtx, cancel := context.WithTimeout(ctx, s.config.ShutdownTimeout)
    defer cancel()
    
    if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
        return fmt.Errorf("failed to shutdown HTTP server: %w", err)
    }
    
    s.logger.Info(ctx, "HTTP server stopped")
    return nil
}

func (s *Server) healthCheck(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "status":    "healthy",
        "timestamp": time.Now().UTC(),
        "version":   "1.0.0",
    })
}

func (s *Server) readinessCheck(c *gin.Context) {
    // Add your readiness checks here (database, cache, etc.)
    c.JSON(http.StatusOK, gin.H{
        "status":    "ready",
        "timestamp": time.Now().UTC(),
    })
}

func (s *Server) metricsHandler(c *gin.Context) {
    // Prometheus metrics endpoint
    // Implementation depends on your metrics library
    c.String(http.StatusOK, "# Metrics endpoint")
}
```

---

## 🔄 **8. BOOTSTRAP SYSTEM TEMPLATE**

### Dependency Injection (src/bootstrap/app.go)
```go
package bootstrap

import (
    "context"
    "os"
    "os/signal"
    "syscall"
    
    "go.uber.org/fx"
    
    "my-backend-system/src/common/configs"
    "my-backend-system/src/common/logger"
    "my-backend-system/src/present/http"
)

func NewApp(configPath string) *fx.App {
    return fx.New(
        // Configuration
        fx.Provide(func() (*configs.Config, error) {
            return configs.LoadConfig(configPath)
        }),
        
        // Logging
        fx.Provide(func(config *configs.Config) logger.Logger {
            if config.App.Debug {
                return logger.NewConsoleLogger(config.Logging.Level)
            }
            return logger.NewLogger(os.Stdout, config.Logging.Level)
        }),
        
        // Database
        ProvideDatabases,
        
        // Cache
        ProvideCache,
        
        // Security
        ProvideSecurity,
        
        // External services
        ProvideExternalServices,
        
        // Repositories
        ProvideRepositories,
        
        // Use cases
        ProvideUseCases,
        
        // HTTP layer
        ProvideHTTPLayer,
        
        // Application lifecycle
        fx.Invoke(func(lc fx.Lifecycle, server *http.Server, logger logger.Logger) {
            lc.Append(fx.Hook{
                OnStart: func(ctx context.Context) error {
                    go func() {
                        if err := server.Start(ctx); err != nil {
                            logger.Fatal(ctx, err, "Failed to start server")
                        }
                    }()
                    return nil
                },
                OnStop: func(ctx context.Context) error {
                    return server.Stop(ctx)
                },
            })
        }),
    )
}

func RunApp(app *fx.App) {
    // Setup signal handling
    ctx, cancel := signal.NotifyContext(context.Background(), 
        os.Interrupt, syscall.SIGTERM)
    defer cancel()
    
    // Start application
    if err := app.Start(ctx); err != nil {
        panic(err)
    }
    
    // Wait for shutdown signal
    <-ctx.Done()
    
    // Stop application
    if err := app.Stop(context.Background()); err != nil {
        panic(err)
    }
}
```

---

## 📝 **9. USAGE EXAMPLE**

### Main Application (main.go)
```go
package main

import (
    "flag"
    
    "my-backend-system/src/bootstrap"
)

func main() {
    var configPath string
    flag.StringVar(&configPath, "config", "configs/config.yaml", "Configuration file path")
    flag.Parse()
    
    app := bootstrap.NewApp(configPath)
    bootstrap.RunApp(app)
}
```

### Docker Configuration
```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/main .
COPY --from=builder /app/configs ./configs

CMD ["./main"]
```

### Docker Compose for Development
```yaml
version: '3.8'

services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      - CONFIG_PATH=configs/config.yaml
    depends_on:
      - postgres
      - redis
    volumes:
      - ./configs:/root/configs

  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: myapp_db
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: password
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data

volumes:
  postgres_data:
  redis_data:
```

---

## 🎯 **KẾT LUẬN**

Template này cung cấp:

1. **✅ Cấu trúc project chuẩn** - Clean Architecture
2. **✅ Configuration management** - Environment-based config
3. **✅ Logging system** - Structured logging với context
4. **✅ Authentication** - JWT-based auth với middleware
5. **✅ Database layer** - Repository pattern với connection pooling
6. **✅ Cache system** - Redis với fallback strategies
7. **✅ HTTP server** - Gin với middleware stack đầy đủ
8. **✅ Dependency injection** - Uber FX cho lifecycle management
9. **✅ Error handling** - Structured error với proper HTTP codes
10. **✅ Testing setup** - Unit, integration test templates

**Cách sử dụng template:**
1. Copy toàn bộ cấu trúc code
2. Thay đổi module name và package names
3. Customize business logic trong `src/core/`
4. Thêm domain-specific repositories và use cases
5. Adjust configuration cho environment cụ thể

**Business Logic Customization:**
- Chỉ cần thay đổi `src/core/domains/`, `src/core/usecases/`
- Thêm domain-specific repositories trong `src/infra/`
- Customize API endpoints trong `src/present/http/`
- Infrastructure layer có thể tái sử dụng 100% 