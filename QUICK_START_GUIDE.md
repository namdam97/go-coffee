# QUICK START GUIDE
## Hướng Dẫn Nhanh Tạo Backend System Từ Template

> **Mục tiêu**: Trong 30 phút, bạn có thể tạo một backend system hoàn chỉnh với đầy đủ tính năng cơ bản.

---

## 🚀 **BƯỚC 1: SETUP PROJECT (5 phút)**

### 1.1 Tạo Project Structure
```bash
# Tạo project mới
mkdir my-awesome-backend
cd my-awesome-backend

# Khởi tạo Go module
go mod init my-awesome-backend

# Tạo cấu trúc thư mục
mkdir -p {configs,src/{bootstrap,common,core,infra,present},tests,scripts,docs}
mkdir -p src/common/{cache,configs,crypto,logger,storage,tracer,fault}
mkdir -p src/core/{domains,enums,usecases}
mkdir -p src/infra/{database,external,cache}
mkdir -p src/present/{http,consumers}
mkdir -p src/present/http/{controllers,middlewares,routes}

echo "✅ Project structure created!"
```

### 1.2 Copy Essential Files
```bash
# Copy từ template
cp BACKEND_CHECKLIST.md my-awesome-backend/
cp IMPLEMENTATION_TEMPLATE.md my-awesome-backend/
cp -r src/* my-awesome-backend/src/
cp configs/* my-awesome-backend/configs/
cp main.go my-awesome-backend/
cp go.mod my-awesome-backend/ # Update module name

echo "✅ Template files copied!"
```

### 1.3 Initialize Dependencies
```bash
cd my-awesome-backend

# Add essential dependencies
go mod tidy
go get github.com/gin-gonic/gin
go get go.uber.org/fx
go get github.com/spf13/viper
go get github.com/rs/zerolog
go get github.com/redis/go-redis/v9
go get github.com/lib/pq

echo "✅ Dependencies installed!"
```

---

## ⚙️ **BƯỚC 2: CONFIGURATION (5 phút)**

### 2.1 Update Config File (configs/config.yaml)
```yaml
app:
  name: "my-awesome-backend"
  version: "1.0.0"
  environment: "development"
  debug: true

server:
  host: "0.0.0.0"
  port: 8080
  read_timeout: "30s"
  write_timeout: "30s"

database:
  type: "postgres"
  host: "localhost"
  port: 5432
  name: "myapp_db"
  username: "postgres"
  password: "password"
  ssl_mode: "disable"

cache:
  redis:
    host: "localhost"
    port: 6379
    password: ""
    db: 0

security:
  jwt:
    access_secret: "your-super-secret-key-change-this-in-production"
    refresh_secret: "your-refresh-secret-key-change-this-too"
    access_expire: "15m"
    refresh_expire: "7d"

logging:
  level: "info"
  format: "json"
```

### 2.2 Environment Setup
```bash
# Tạo .env file cho development
cat > .env << EOF
DB_HOST=localhost
DB_PORT=5432
DB_NAME=myapp_db
DB_USER=postgres
DB_PASSWORD=password
REDIS_HOST=localhost
REDIS_PORT=6379
JWT_SECRET=your-super-secret-key
EOF

echo "✅ Configuration ready!"
```

---

## 🏗️ **BƯỚC 3: BUSINESS LOGIC CUSTOMIZATION (10 phút)**

### 3.1 Define Your Domain (Example: E-commerce)
```go
// src/core/domains/product.go
package domains

import (
    "context"
    "time"
)

type Product struct {
    ID          string    `json:"id" db:"id"`
    Name        string    `json:"name" db:"name"`
    Description string    `json:"description" db:"description"`
    Price       float64   `json:"price" db:"price"`
    Stock       int       `json:"stock" db:"stock"`
    CategoryID  string    `json:"category_id" db:"category_id"`
    CreatedAt   time.Time `json:"created_at" db:"created_at"`
    UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type ProductRepository interface {
    Create(ctx context.Context, product *Product) error
    GetByID(ctx context.Context, id string) (*Product, error)
    List(ctx context.Context, limit, offset int) ([]*Product, error)
    Update(ctx context.Context, product *Product) error
    Delete(ctx context.Context, id string) error
}
```

```go
// src/core/domains/order.go
package domains

import (
    "context"
    "time"
)

type Order struct {
    ID         string      `json:"id" db:"id"`
    UserID     string      `json:"user_id" db:"user_id"`
    Items      []OrderItem `json:"items"`
    TotalPrice float64     `json:"total_price" db:"total_price"`
    Status     string      `json:"status" db:"status"`
    CreatedAt  time.Time   `json:"created_at" db:"created_at"`
}

type OrderItem struct {
    ProductID string  `json:"product_id"`
    Quantity  int     `json:"quantity"`
    Price     float64 `json:"price"`
}

type OrderRepository interface {
    Create(ctx context.Context, order *Order) error
    GetByID(ctx context.Context, id string) (*Order, error)
    GetByUserID(ctx context.Context, userID string) ([]*Order, error)
    UpdateStatus(ctx context.Context, id, status string) error
}
```

### 3.2 Create Use Cases
```go
// src/core/usecases/product_usecase.go
package usecases

import (
    "context"
    "fmt"
    
    "my-awesome-backend/src/core/domains"
    "my-awesome-backend/src/common/logger"
)

type ProductUseCase struct {
    productRepo domains.ProductRepository
    logger      logger.Logger
}

func NewProductUseCase(productRepo domains.ProductRepository, logger logger.Logger) *ProductUseCase {
    return &ProductUseCase{
        productRepo: productRepo,
        logger:      logger,
    }
}

func (uc *ProductUseCase) CreateProduct(ctx context.Context, product *domains.Product) error {
    // Business validation
    if product.Price <= 0 {
        return fmt.Errorf("product price must be greater than 0")
    }
    
    if product.Stock < 0 {
        return fmt.Errorf("product stock cannot be negative")
    }
    
    // Generate ID
    product.ID = generateID()
    
    // Save to repository
    if err := uc.productRepo.Create(ctx, product); err != nil {
        uc.logger.Error(ctx, err, "Failed to create product")
        return err
    }
    
    uc.logger.Info(ctx, "Product created successfully", 
        logger.String("product_id", product.ID))
    
    return nil
}

func (uc *ProductUseCase) GetProduct(ctx context.Context, id string) (*domains.Product, error) {
    product, err := uc.productRepo.GetByID(ctx, id)
    if err != nil {
        uc.logger.Error(ctx, err, "Failed to get product", 
            logger.String("product_id", id))
        return nil, err
    }
    
    return product, nil
}

func generateID() string {
    // Implementation for ID generation
    return "prod_" + randomString(10)
}
```

### 3.3 Create Controllers
```go
// src/present/http/controllers/product_controller.go
package controllers

import (
    "net/http"
    
    "github.com/gin-gonic/gin"
    
    "my-awesome-backend/src/core/domains"
    "my-awesome-backend/src/core/usecases"
    "my-awesome-backend/src/common/logger"
)

type ProductController struct {
    productUC *usecases.ProductUseCase
    logger    logger.Logger
}

func NewProductController(productUC *usecases.ProductUseCase, logger logger.Logger) *ProductController {
    return &ProductController{
        productUC: productUC,
        logger:    logger,
    }
}

func (ctrl *ProductController) CreateProduct(c *gin.Context) {
    var product domains.Product
    if err := c.ShouldBindJSON(&product); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    if err := ctrl.productUC.CreateProduct(c.Request.Context(), &product); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusCreated, gin.H{
        "message": "Product created successfully",
        "data":    product,
    })
}

func (ctrl *ProductController) GetProduct(c *gin.Context) {
    id := c.Param("id")
    
    product, err := ctrl.productUC.GetProduct(c.Request.Context(), id)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "data": product,
    })
}
```

---

## 🔗 **BƯỚC 4: SETUP ROUTES (3 phút)**

### 4.1 Create Route Groups
```go
// src/present/http/routes/product_routes.go
package routes

import (
    "github.com/gin-gonic/gin"
    
    "my-awesome-backend/src/present/http/controllers"
    "my-awesome-backend/src/present/http/middlewares"
)

type ProductRoutes struct {
    productController *controllers.ProductController
    authMiddleware    *middlewares.AuthMiddleware
}

func NewProductRoutes(
    productController *controllers.ProductController,
    authMiddleware *middlewares.AuthMiddleware,
) *ProductRoutes {
    return &ProductRoutes{
        productController: productController,
        authMiddleware:    authMiddleware,
    }
}

func (r *ProductRoutes) Setup(router *gin.RouterGroup) {
    products := router.Group("/products")
    {
        products.POST("", r.authMiddleware.RequireAuth(), r.productController.CreateProduct)
        products.GET("/:id", r.productController.GetProduct)
        products.GET("", r.productController.ListProducts)
        products.PUT("/:id", r.authMiddleware.RequireAuth(), r.productController.UpdateProduct)
        products.DELETE("/:id", r.authMiddleware.RequireAuth(), r.productController.DeleteProduct)
    }
}
```

### 4.2 Register Routes in Bootstrap
```go
// src/bootstrap/http.go
func ProvideHTTPLayer(
    config *configs.Config,
    logger logger.Logger,
    productController *controllers.ProductController,
    authMiddleware *middlewares.AuthMiddleware,
) *http.Server {
    server := http.NewServer(&config.Server, logger)
    
    // Setup middlewares
    server.SetupMiddlewares()
    
    // Setup routes
    productRoutes := routes.NewProductRoutes(productController, authMiddleware)
    server.SetupRoutes(productRoutes)
    
    return server
}
```

---

## 🚀 **BƯỚC 5: RUN & TEST (5 phút)**

### 5.1 Start Dependencies
```bash
# Start PostgreSQL và Redis với Docker
docker run -d --name postgres \
  -e POSTGRES_DB=myapp_db \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=password \
  -p 5432:5432 postgres:15

docker run -d --name redis \
  -p 6379:6379 redis:7-alpine

echo "✅ Dependencies started!"
```

### 5.2 Run Application
```bash
# Build and run
go build -o app .
./app -config configs/config.yaml

# Or run directly
go run main.go -config configs/config.yaml

echo "🚀 Server running on http://localhost:8080"
```

### 5.3 Test APIs
```bash
# Health check
curl http://localhost:8080/health

# Create product (need auth token)
curl -X POST http://localhost:8080/api/products \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "name": "iPhone 15",
    "description": "Latest iPhone model",
    "price": 999.99,
    "stock": 100,
    "category_id": "electronics"
  }'

# Get product
curl http://localhost:8080/api/products/prod_123456

echo "✅ APIs working!"
```

---

## 📋 **BƯỚC 6: PRODUCTION READY (2 phút)**

### 6.1 Create Dockerfile
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
EXPOSE 8080
CMD ["./main"]
```

### 6.2 Create Docker Compose
```yaml
version: '3.8'
services:
  app:
    build: .
    ports:
      - "8080:8080"
    depends_on:
      - postgres
      - redis
    environment:
      - CONFIG_PATH=configs/config.yaml

  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: myapp_db
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: password
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    volumes:
      - redis_data:/data

volumes:
  postgres_data:
  redis_data:
```

### 6.3 Deploy
```bash
# Build and run with Docker Compose
docker-compose up -d

echo "🎉 Production ready!"
```

---

## 🎯 **CUSTOMIZATION CHECKLIST**

### ✅ Business Logic Customization
- [ ] **Domain Models** - Define your business entities
- [ ] **Repository Interfaces** - Data access contracts
- [ ] **Use Cases** - Business logic implementation
- [ ] **Controllers** - HTTP request handlers
- [ ] **Routes** - API endpoint definitions
- [ ] **Validation** - Input validation rules
- [ ] **Business Rules** - Domain-specific constraints

### ✅ Infrastructure Customization
- [ ] **Database Schema** - Create tables for your domain
- [ ] **Migration Scripts** - Database schema changes
- [ ] **External APIs** - Third-party integrations
- [ ] **Background Jobs** - Async processing
- [ ] **Notifications** - Email, SMS, push notifications
- [ ] **File Storage** - Upload/download functionality

### ✅ Configuration Customization
- [ ] **Environment Variables** - Production settings
- [ ] **Feature Flags** - Toggle features on/off
- [ ] **Rate Limits** - API throttling configuration
- [ ] **Cache TTL** - Cache expiration settings
- [ ] **Security Settings** - JWT expiration, API keys
- [ ] **External Service URLs** - Third-party endpoints

---

## 🛠️ **DEVELOPMENT WORKFLOW**

### Daily Development
```bash
# 1. Start dependencies
docker-compose up -d postgres redis

# 2. Run application
go run main.go

# 3. Test changes
curl http://localhost:8080/health

# 4. Run tests
go test ./...

# 5. Build for production
go build -o app .
```

### Adding New Features
```bash
# 1. Create domain model
# src/core/domains/new_feature.go

# 2. Create repository interface
# src/core/domains/new_feature_repo.go

# 3. Implement repository
# src/infra/database/new_feature_repo.go

# 4. Create use case
# src/core/usecases/new_feature_usecase.go

# 5. Create controller
# src/present/http/controllers/new_feature_controller.go

# 6. Add routes
# src/present/http/routes/new_feature_routes.go

# 7. Update bootstrap
# src/bootstrap/providers.go
```

---

## 📚 **NEXT STEPS**

### Immediate (Week 1)
- [ ] Add database migrations
- [ ] Implement user authentication
- [ ] Add input validation
- [ ] Create unit tests
- [ ] Setup logging

### Short-term (Week 2-4)
- [ ] Add caching layer
- [ ] Implement rate limiting
- [ ] Add monitoring/metrics
- [ ] Setup CI/CD pipeline
- [ ] Add integration tests

### Long-term (Month 2+)
- [ ] Microservices decomposition
- [ ] Event-driven architecture
- [ ] Advanced security features
- [ ] Performance optimization
- [ ] Scalability improvements

---

## 🆘 **TROUBLESHOOTING**

### Common Issues

**1. Database Connection Failed**
```bash
# Check if PostgreSQL is running
docker ps | grep postgres

# Check connection
psql -h localhost -U postgres -d myapp_db
```

**2. Redis Connection Failed**
```bash
# Check if Redis is running
docker ps | grep redis

# Test connection
redis-cli -h localhost -p 6379 ping
```

**3. Port Already in Use**
```bash
# Find process using port 8080
lsof -i :8080

# Kill process
kill -9 <PID>
```

**4. Module Import Issues**
```bash
# Update module name in go.mod
go mod edit -module your-new-module-name

# Update imports in all files
find . -name "*.go" -exec sed -i 's/my-awesome-backend/your-new-module-name/g' {} \;
```

---

## 🎉 **CONGRATULATIONS!**

Trong 30 phút, bạn đã có:
- ✅ **Hoàn chỉnh backend system** với đầy đủ tính năng cơ bản
- ✅ **Clean Architecture** với separation of concerns  
- ✅ **Production-ready** với Docker, monitoring, logging
- ✅ **Scalable foundation** để mở rộng business logic
- ✅ **Best practices** được tích hợp sẵn

**Giờ bạn chỉ cần tập trung vào business logic của domain cụ thể!**

---

## 📞 **SUPPORT**

Nếu gặp vấn đề:
1. Check `BACKEND_CHECKLIST.md` cho requirements đầy đủ
2. Xem `IMPLEMENTATION_TEMPLATE.md` cho code examples chi tiết  
3. Tham khảo core-service như reference implementation
4. Google/StackOverflow cho specific technical issues

**Happy Coding! 🚀** 