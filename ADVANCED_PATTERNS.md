# ADVANCED BACKEND PATTERNS & BEST PRACTICES
## Mở Rộng Từ Core-Service Template Cho Enterprise Systems

> **Mục tiêu**: Đào sâu các patterns nâng cao, microservices, event-driven architecture, và enterprise-grade features dựa trên foundation đã xây dựng.

---

## 🏗️ **1. ADVANCED ARCHITECTURAL PATTERNS**

### 🔄 **Event-Driven Architecture (EDA)**

#### **1.1 Event Sourcing Pattern**
```go
// src/common/eventsourcing/event_store.go
package eventsourcing

import (
    "context"
    "encoding/json"
    "time"
)

type Event struct {
    ID            string                 `json:"id"`
    AggregateID   string                 `json:"aggregate_id"`
    AggregateType string                 `json:"aggregate_type"`
    EventType     string                 `json:"event_type"`
    EventData     map[string]interface{} `json:"event_data"`
    Version       int64                  `json:"version"`
    Timestamp     time.Time              `json:"timestamp"`
    Metadata      map[string]string      `json:"metadata"`
}

type EventStore interface {
    // Write events
    SaveEvents(ctx context.Context, aggregateID string, events []Event, expectedVersion int64) error
    
    // Read events
    GetEvents(ctx context.Context, aggregateID string, fromVersion int64) ([]Event, error)
    GetEventsByType(ctx context.Context, eventType string, from, to time.Time) ([]Event, error)
    
    // Snapshots for performance
    SaveSnapshot(ctx context.Context, aggregateID string, snapshot Snapshot) error
    GetSnapshot(ctx context.Context, aggregateID string) (*Snapshot, error)
    
    // Event streaming
    SubscribeToEvents(ctx context.Context, eventTypes []string) (<-chan Event, error)
}

type Snapshot struct {
    AggregateID   string                 `json:"aggregate_id"`
    AggregateType string                 `json:"aggregate_type"`
    Data          map[string]interface{} `json:"data"`
    Version       int64                  `json:"version"`
    Timestamp     time.Time              `json:"timestamp"`
}

// Aggregate Root Pattern
type AggregateRoot interface {
    GetID() string
    GetVersion() int64
    GetUncommittedEvents() []Event
    MarkEventsAsCommitted()
    LoadFromHistory(events []Event)
}

// Example: Order Aggregate
type OrderAggregate struct {
    id               string
    customerID       string
    items            []OrderItem
    status           string
    totalAmount      float64
    version          int64
    uncommittedEvents []Event
}

func (o *OrderAggregate) CreateOrder(customerID string, items []OrderItem) error {
    // Business validation
    if len(items) == 0 {
        return errors.New("order must have at least one item")
    }
    
    // Calculate total
    total := 0.0
    for _, item := range items {
        total += item.Price * float64(item.Quantity)
    }
    
    // Create event
    event := Event{
        ID:            generateEventID(),
        AggregateID:   o.id,
        AggregateType: "Order",
        EventType:     "OrderCreated",
        EventData: map[string]interface{}{
            "customer_id":  customerID,
            "items":        items,
            "total_amount": total,
            "status":       "pending",
        },
        Version:   o.version + 1,
        Timestamp: time.Now(),
    }
    
    o.applyEvent(event)
    return nil
}

func (o *OrderAggregate) applyEvent(event Event) {
    switch event.EventType {
    case "OrderCreated":
        o.customerID = event.EventData["customer_id"].(string)
        o.items = event.EventData["items"].([]OrderItem)
        o.totalAmount = event.EventData["total_amount"].(float64)
        o.status = event.EventData["status"].(string)
    case "OrderStatusChanged":
        o.status = event.EventData["new_status"].(string)
    }
    
    o.version = event.Version
    o.uncommittedEvents = append(o.uncommittedEvents, event)
}
```

#### **1.2 CQRS (Command Query Responsibility Segregation)**
```go
// src/common/cqrs/command_bus.go
package cqrs

import (
    "context"
    "fmt"
    "reflect"
)

type Command interface {
    GetID() string
    GetType() string
}

type CommandHandler interface {
    Handle(ctx context.Context, cmd Command) error
}

type CommandBus interface {
    Send(ctx context.Context, cmd Command) error
    Register(cmdType string, handler CommandHandler)
}

type InMemoryCommandBus struct {
    handlers map[string]CommandHandler
    logger   logger.Logger
}

func NewInMemoryCommandBus(logger logger.Logger) *InMemoryCommandBus {
    return &InMemoryCommandBus{
        handlers: make(map[string]CommandHandler),
        logger:   logger,
    }
}

func (bus *InMemoryCommandBus) Send(ctx context.Context, cmd Command) error {
    handler, exists := bus.handlers[cmd.GetType()]
    if !exists {
        return fmt.Errorf("no handler registered for command type: %s", cmd.GetType())
    }
    
    bus.logger.Info(ctx, "Executing command",
        logger.String("command_id", cmd.GetID()),
        logger.String("command_type", cmd.GetType()))
    
    if err := handler.Handle(ctx, cmd); err != nil {
        bus.logger.Error(ctx, err, "Command execution failed",
            logger.String("command_id", cmd.GetID()),
            logger.String("command_type", cmd.GetType()))
        return err
    }
    
    bus.logger.Info(ctx, "Command executed successfully",
        logger.String("command_id", cmd.GetID()),
        logger.String("command_type", cmd.GetType()))
    
    return nil
}

func (bus *InMemoryCommandBus) Register(cmdType string, handler CommandHandler) {
    bus.handlers[cmdType] = handler
}

// Query Bus
type Query interface {
    GetID() string
    GetType() string
}

type QueryHandler interface {
    Handle(ctx context.Context, query Query) (interface{}, error)
}

type QueryBus interface {
    Send(ctx context.Context, query Query) (interface{}, error)
    Register(queryType string, handler QueryHandler)
}

// Example: Create Order Command
type CreateOrderCommand struct {
    ID         string      `json:"id"`
    CustomerID string      `json:"customer_id"`
    Items      []OrderItem `json:"items"`
}

func (c CreateOrderCommand) GetID() string   { return c.ID }
func (c CreateOrderCommand) GetType() string { return "CreateOrder" }

type CreateOrderCommandHandler struct {
    orderRepo   domains.OrderRepository
    eventStore  eventsourcing.EventStore
    logger      logger.Logger
}

func (h *CreateOrderCommandHandler) Handle(ctx context.Context, cmd Command) error {
    createOrderCmd := cmd.(CreateOrderCommand)
    
    // Load or create aggregate
    aggregate := &OrderAggregate{id: createOrderCmd.ID}
    
    // Execute business logic
    if err := aggregate.CreateOrder(createOrderCmd.CustomerID, createOrderCmd.Items); err != nil {
        return err
    }
    
    // Save events
    events := aggregate.GetUncommittedEvents()
    if err := h.eventStore.SaveEvents(ctx, aggregate.GetID(), events, aggregate.GetVersion()-int64(len(events))); err != nil {
        return err
    }
    
    aggregate.MarkEventsAsCommitted()
    return nil
}
```

#### **1.3 Saga Pattern (Distributed Transactions)**
```go
// src/common/saga/saga.go
package saga

import (
    "context"
    "time"
)

type SagaStep struct {
    Name         string
    Execute      func(ctx context.Context, data map[string]interface{}) error
    Compensate   func(ctx context.Context, data map[string]interface{}) error
    RetryPolicy  *RetryPolicy
}

type RetryPolicy struct {
    MaxAttempts int
    Delay       time.Duration
    Backoff     float64
}

type SagaDefinition struct {
    Name  string
    Steps []SagaStep
}

type SagaExecution struct {
    ID           string                 `json:"id"`
    SagaName     string                 `json:"saga_name"`
    Data         map[string]interface{} `json:"data"`
    CurrentStep  int                    `json:"current_step"`
    Status       string                 `json:"status"` // running, completed, failed, compensating, compensated
    ExecutedSteps []string              `json:"executed_steps"`
    CreatedAt    time.Time              `json:"created_at"`
    UpdatedAt    time.Time              `json:"updated_at"`
}

type SagaOrchestrator interface {
    StartSaga(ctx context.Context, sagaName string, data map[string]interface{}) (*SagaExecution, error)
    ContinueSaga(ctx context.Context, executionID string) error
    CompensateSaga(ctx context.Context, executionID string) error
    GetSagaStatus(ctx context.Context, executionID string) (*SagaExecution, error)
}

// Example: Order Processing Saga
func CreateOrderProcessingSaga() *SagaDefinition {
    return &SagaDefinition{
        Name: "OrderProcessing",
        Steps: []SagaStep{
            {
                Name: "ValidatePayment",
                Execute: func(ctx context.Context, data map[string]interface{}) error {
                    // Call payment service to validate
                    paymentService := getPaymentService()
                    return paymentService.ValidatePayment(ctx, data["payment_info"])
                },
                Compensate: func(ctx context.Context, data map[string]interface{}) error {
                    // Release payment hold
                    paymentService := getPaymentService()
                    return paymentService.ReleaseHold(ctx, data["payment_id"])
                },
            },
            {
                Name: "ReserveInventory",
                Execute: func(ctx context.Context, data map[string]interface{}) error {
                    // Reserve inventory
                    inventoryService := getInventoryService()
                    return inventoryService.ReserveItems(ctx, data["items"])
                },
                Compensate: func(ctx context.Context, data map[string]interface{}) error {
                    // Release inventory reservation
                    inventoryService := getInventoryService()
                    return inventoryService.ReleaseReservation(ctx, data["reservation_id"])
                },
            },
            {
                Name: "ChargePayment",
                Execute: func(ctx context.Context, data map[string]interface{}) error {
                    // Charge payment
                    paymentService := getPaymentService()
                    return paymentService.ChargePayment(ctx, data["payment_info"])
                },
                Compensate: func(ctx context.Context, data map[string]interface{}) error {
                    // Refund payment
                    paymentService := getPaymentService()
                    return paymentService.RefundPayment(ctx, data["payment_id"])
                },
            },
            {
                Name: "CreateShipment",
                Execute: func(ctx context.Context, data map[string]interface{}) error {
                    // Create shipment
                    shippingService := getShippingService()
                    return shippingService.CreateShipment(ctx, data["order_info"])
                },
                Compensate: func(ctx context.Context, data map[string]interface{}) error {
                    // Cancel shipment
                    shippingService := getShippingService()
                    return shippingService.CancelShipment(ctx, data["shipment_id"])
                },
            },
        },
    }
}
```

---

## 🔄 **2. MICROSERVICES PATTERNS**

### **2.1 Service Discovery & Registry**
```go
// src/common/discovery/service_registry.go
package discovery

import (
    "context"
    "encoding/json"
    "fmt"
    "time"
)

type ServiceInstance struct {
    ID       string            `json:"id"`
    Name     string            `json:"name"`
    Version  string            `json:"version"`
    Host     string            `json:"host"`
    Port     int               `json:"port"`
    Tags     []string          `json:"tags"`
    Metadata map[string]string `json:"metadata"`
    Health   HealthCheck       `json:"health"`
}

type HealthCheck struct {
    Endpoint string        `json:"endpoint"`
    Interval time.Duration `json:"interval"`
    Timeout  time.Duration `json:"timeout"`
}

type ServiceRegistry interface {
    Register(ctx context.Context, instance *ServiceInstance) error
    Deregister(ctx context.Context, instanceID string) error
    Discover(ctx context.Context, serviceName string) ([]*ServiceInstance, error)
    Watch(ctx context.Context, serviceName string) (<-chan []*ServiceInstance, error)
    HealthCheck(ctx context.Context, instanceID string) error
}

// Consul-based implementation
type ConsulServiceRegistry struct {
    client consulapi.Client
    logger logger.Logger
}

func (r *ConsulServiceRegistry) Register(ctx context.Context, instance *ServiceInstance) error {
    registration := &consulapi.AgentServiceRegistration{
        ID:      instance.ID,
        Name:    instance.Name,
        Tags:    instance.Tags,
        Port:    instance.Port,
        Address: instance.Host,
        Meta:    instance.Metadata,
        Check: &consulapi.AgentServiceCheck{
            HTTP:                           fmt.Sprintf("http://%s:%d%s", instance.Host, instance.Port, instance.Health.Endpoint),
            Interval:                       instance.Health.Interval.String(),
            Timeout:                        instance.Health.Timeout.String(),
            DeregisterCriticalServiceAfter: "30s",
        },
    }
    
    return r.client.Agent().ServiceRegister(registration)
}

// Load Balancer
type LoadBalancer interface {
    SelectInstance(instances []*ServiceInstance) (*ServiceInstance, error)
}

type RoundRobinLoadBalancer struct {
    counter int64
}

func (lb *RoundRobinLoadBalancer) SelectInstance(instances []*ServiceInstance) (*ServiceInstance, error) {
    if len(instances) == 0 {
        return nil, errors.New("no available instances")
    }
    
    index := atomic.AddInt64(&lb.counter, 1) % int64(len(instances))
    return instances[index], nil
}

// Service Client with Discovery
type ServiceClient struct {
    serviceName string
    registry    ServiceRegistry
    loadBalancer LoadBalancer
    httpClient  *http.Client
    logger      logger.Logger
}

func (c *ServiceClient) Call(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
    // Discover service instances
    instances, err := c.registry.Discover(ctx, c.serviceName)
    if err != nil {
        return nil, fmt.Errorf("service discovery failed: %w", err)
    }
    
    // Select instance using load balancer
    instance, err := c.loadBalancer.SelectInstance(instances)
    if err != nil {
        return nil, fmt.Errorf("load balancing failed: %w", err)
    }
    
    // Make HTTP call
    url := fmt.Sprintf("http://%s:%d%s", instance.Host, instance.Port, path)
    
    var req *http.Request
    if body != nil {
        jsonBody, _ := json.Marshal(body)
        req, err = http.NewRequestWithContext(ctx, method, url, strings.NewReader(string(jsonBody)))
    } else {
        req, err = http.NewRequestWithContext(ctx, method, url, nil)
    }
    
    if err != nil {
        return nil, err
    }
    
    req.Header.Set("Content-Type", "application/json")
    
    return c.httpClient.Do(req)
}
```

### **2.2 Circuit Breaker Pattern**
```go
// src/common/circuitbreaker/circuit_breaker.go
package circuitbreaker

import (
    "context"
    "errors"
    "sync"
    "time"
)

type State int

const (
    StateClosed State = iota
    StateHalfOpen
    StateOpen
)

type CircuitBreaker struct {
    name           string
    maxRequests    uint32
    interval       time.Duration
    timeout        time.Duration
    failureRatio   float64
    
    mutex          sync.RWMutex
    state          State
    generation     uint64
    counts         Counts
    expiry         time.Time
    
    onStateChange  func(name string, from State, to State)
}

type Counts struct {
    Requests             uint32
    TotalSuccesses       uint32
    TotalFailures        uint32
    ConsecutiveSuccesses uint32
    ConsecutiveFailures  uint32
}

func NewCircuitBreaker(settings Settings) *CircuitBreaker {
    cb := &CircuitBreaker{
        name:          settings.Name,
        maxRequests:   settings.MaxRequests,
        interval:      settings.Interval,
        timeout:       settings.Timeout,
        failureRatio:  settings.FailureRatio,
        onStateChange: settings.OnStateChange,
    }
    
    cb.toNewGeneration(time.Now())
    return cb
}

func (cb *CircuitBreaker) Execute(req func() (interface{}, error)) (interface{}, error) {
    generation, err := cb.beforeRequest()
    if err != nil {
        return nil, err
    }
    
    defer func() {
        e := recover()
        if e != nil {
            cb.afterRequest(generation, false)
            panic(e)
        }
    }()
    
    result, err := req()
    cb.afterRequest(generation, err == nil)
    return result, err
}

func (cb *CircuitBreaker) beforeRequest() (uint64, error) {
    cb.mutex.Lock()
    defer cb.mutex.Unlock()
    
    now := time.Now()
    state, generation := cb.currentState(now)
    
    if state == StateOpen {
        return generation, errors.New("circuit breaker is open")
    } else if state == StateHalfOpen && cb.counts.Requests >= cb.maxRequests {
        return generation, errors.New("circuit breaker is half-open and max requests reached")
    }
    
    cb.counts.onRequest()
    return generation, nil
}

func (cb *CircuitBreaker) afterRequest(before uint64, success bool) {
    cb.mutex.Lock()
    defer cb.mutex.Unlock()
    
    now := time.Now()
    state, generation := cb.currentState(now)
    if generation != before {
        return
    }
    
    if success {
        cb.onSuccess(state, now)
    } else {
        cb.onFailure(state, now)
    }
}

// Rate Limiter
type RateLimiter interface {
    Allow() bool
    Wait(ctx context.Context) error
}

type TokenBucketRateLimiter struct {
    rate     float64
    capacity int64
    tokens   int64
    lastTime time.Time
    mutex    sync.Mutex
}

func NewTokenBucketRateLimiter(rate float64, capacity int64) *TokenBucketRateLimiter {
    return &TokenBucketRateLimiter{
        rate:     rate,
        capacity: capacity,
        tokens:   capacity,
        lastTime: time.Now(),
    }
}

func (rl *TokenBucketRateLimiter) Allow() bool {
    rl.mutex.Lock()
    defer rl.mutex.Unlock()
    
    now := time.Now()
    elapsed := now.Sub(rl.lastTime).Seconds()
    
    // Add tokens based on elapsed time
    rl.tokens += int64(elapsed * rl.rate)
    if rl.tokens > rl.capacity {
        rl.tokens = rl.capacity
    }
    
    rl.lastTime = now
    
    if rl.tokens > 0 {
        rl.tokens--
        return true
    }
    
    return false
}
```

### **2.3 API Gateway Pattern**
```go
// src/gateway/gateway.go
package gateway

import (
    "context"
    "net/http"
    "net/http/httputil"
    "net/url"
    "strings"
    "time"
)

type Route struct {
    Path        string
    Method      string
    ServiceName string
    Rewrite     string
    Middleware  []MiddlewareFunc
    RateLimit   *RateLimitConfig
    Auth        *AuthConfig
    Timeout     time.Duration
}

type RateLimitConfig struct {
    RequestsPerSecond int
    BurstSize         int
}

type AuthConfig struct {
    Required bool
    Roles    []string
    Scopes   []string
}

type APIGateway struct {
    routes          []Route
    serviceRegistry discovery.ServiceRegistry
    loadBalancer    discovery.LoadBalancer
    circuitBreaker  map[string]*circuitbreaker.CircuitBreaker
    rateLimiter     map[string]RateLimiter
    logger          logger.Logger
}

func (gw *APIGateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    // Find matching route
    route := gw.findRoute(r.Method, r.URL.Path)
    if route == nil {
        http.Error(w, "Route not found", http.StatusNotFound)
        return
    }
    
    // Apply middleware
    handler := gw.createHandler(route)
    for i := len(route.Middleware) - 1; i >= 0; i-- {
        handler = route.Middleware[i](handler)
    }
    
    // Execute with timeout
    if route.Timeout > 0 {
        var cancel context.CancelFunc
        ctx, cancel = context.WithTimeout(ctx, route.Timeout)
        defer cancel()
        r = r.WithContext(ctx)
    }
    
    handler.ServeHTTP(w, r)
}

func (gw *APIGateway) createHandler(route *Route) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Service discovery
        instances, err := gw.serviceRegistry.Discover(r.Context(), route.ServiceName)
        if err != nil {
            gw.logger.Error(r.Context(), err, "Service discovery failed")
            http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
            return
        }
        
        // Load balancing
        instance, err := gw.loadBalancer.SelectInstance(instances)
        if err != nil {
            gw.logger.Error(r.Context(), err, "Load balancing failed")
            http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
            return
        }
        
        // Circuit breaker
        cb := gw.circuitBreaker[route.ServiceName]
        if cb != nil {
            result, err := cb.Execute(func() (interface{}, error) {
                return gw.proxyRequest(w, r, instance, route)
            })
            
            if err != nil {
                gw.logger.Error(r.Context(), err, "Circuit breaker error")
                http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
                return
            }
            
            // Result is handled in proxyRequest
            return
        }
        
        // Direct proxy without circuit breaker
        gw.proxyRequest(w, r, instance, route)
    })
}

func (gw *APIGateway) proxyRequest(w http.ResponseWriter, r *http.Request, instance *discovery.ServiceInstance, route *Route) (interface{}, error) {
    target, _ := url.Parse(fmt.Sprintf("http://%s:%d", instance.Host, instance.Port))
    
    // Rewrite path if needed
    if route.Rewrite != "" {
        r.URL.Path = strings.Replace(r.URL.Path, route.Path, route.Rewrite, 1)
    }
    
    proxy := httputil.NewSingleHostReverseProxy(target)
    
    // Customize proxy behavior
    proxy.ModifyResponse = func(resp *http.Response) error {
        // Add custom headers
        resp.Header.Set("X-Gateway", "api-gateway")
        resp.Header.Set("X-Service", route.ServiceName)
        return nil
    }
    
    proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
        gw.logger.Error(r.Context(), err, "Proxy error")
        http.Error(w, "Bad gateway", http.StatusBadGateway)
    }
    
    proxy.ServeHTTP(w, r)
    return nil, nil
}

// Gateway Middleware
type MiddlewareFunc func(http.Handler) http.Handler

func AuthMiddleware(authService AuthService) MiddlewareFunc {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            token := extractToken(r)
            if token == "" {
                http.Error(w, "Unauthorized", http.StatusUnauthorized)
                return
            }
            
            claims, err := authService.ValidateToken(r.Context(), token)
            if err != nil {
                http.Error(w, "Invalid token", http.StatusUnauthorized)
                return
            }
            
            // Add claims to context
            ctx := context.WithValue(r.Context(), "claims", claims)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

func RateLimitMiddleware(limiter RateLimiter) MiddlewareFunc {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if !limiter.Allow() {
                http.Error(w, "Too many requests", http.StatusTooManyRequests)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}
```

---

## 📊 **3. ADVANCED MONITORING & OBSERVABILITY**

### **3.1 Distributed Tracing với Custom Spans**
```go
// src/common/tracing/tracer.go
package tracing

import (
    "context"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/trace"
)

type TracingService struct {
    tracer trace.Tracer
}

func NewTracingService(serviceName string) *TracingService {
    tracer := otel.Tracer(serviceName)
    return &TracingService{tracer: tracer}
}

func (ts *TracingService) StartSpan(ctx context.Context, operationName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
    return ts.tracer.Start(ctx, operationName, opts...)
}

// Custom span decorators
func (ts *TracingService) TraceDBOperation(ctx context.Context, operation, table string) (context.Context, trace.Span) {
    return ts.tracer.Start(ctx, fmt.Sprintf("db.%s", operation),
        trace.WithAttributes(
            attribute.String("db.operation", operation),
            attribute.String("db.table", table),
            attribute.String("db.system", "postgresql"),
        ),
        trace.WithSpanKind(trace.SpanKindClient),
    )
}

func (ts *TracingService) TraceHTTPRequest(ctx context.Context, method, url string) (context.Context, trace.Span) {
    return ts.tracer.Start(ctx, fmt.Sprintf("http.%s", method),
        trace.WithAttributes(
            attribute.String("http.method", method),
            attribute.String("http.url", url),
        ),
        trace.WithSpanKind(trace.SpanKindClient),
    )
}

func (ts *TracingService) TraceBusinessOperation(ctx context.Context, domain, operation string, attributes map[string]interface{}) (context.Context, trace.Span) {
    attrs := make([]attribute.KeyValue, 0, len(attributes)+2)
    attrs = append(attrs, 
        attribute.String("business.domain", domain),
        attribute.String("business.operation", operation),
    )
    
    for k, v := range attributes {
        switch val := v.(type) {
        case string:
            attrs = append(attrs, attribute.String(k, val))
        case int:
            attrs = append(attrs, attribute.Int(k, val))
        case int64:
            attrs = append(attrs, attribute.Int64(k, val))
        case float64:
            attrs = append(attrs, attribute.Float64(k, val))
        case bool:
            attrs = append(attrs, attribute.Bool(k, val))
        }
    }
    
    return ts.tracer.Start(ctx, fmt.Sprintf("%s.%s", domain, operation),
        trace.WithAttributes(attrs...),
        trace.WithSpanKind(trace.SpanKindInternal),
    )
}

// Middleware cho automatic tracing
func TracingMiddleware(tracer trace.Tracer) gin.HandlerFunc {
    return func(c *gin.Context) {
        ctx, span := tracer.Start(c.Request.Context(), c.Request.URL.Path,
            trace.WithAttributes(
                attribute.String("http.method", c.Request.Method),
                attribute.String("http.route", c.FullPath()),
                attribute.String("http.user_agent", c.Request.UserAgent()),
            ),
            trace.WithSpanKind(trace.SpanKindServer),
        )
        defer span.End()
        
        c.Request = c.Request.WithContext(ctx)
        c.Next()
        
        // Add response attributes
        span.SetAttributes(
            attribute.Int("http.status_code", c.Writer.Status()),
            attribute.Int("http.response_size", c.Writer.Size()),
        )
        
        if c.Writer.Status() >= 400 {
            span.SetStatus(codes.Error, fmt.Sprintf("HTTP %d", c.Writer.Status()))
        }
    }
}
```

### **3.2 Custom Metrics với Prometheus**
```go
// src/common/metrics/metrics.go
package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

type MetricsService struct {
    // HTTP metrics
    httpRequestsTotal    *prometheus.CounterVec
    httpRequestDuration  *prometheus.HistogramVec
    httpRequestSize      *prometheus.HistogramVec
    httpResponseSize     *prometheus.HistogramVec
    
    // Business metrics
    businessOperationsTotal    *prometheus.CounterVec
    businessOperationDuration *prometheus.HistogramVec
    
    // Database metrics
    dbConnectionsActive *prometheus.GaugeVec
    dbQueryDuration     *prometheus.HistogramVec
    dbTransactionsTotal *prometheus.CounterVec
    
    // Cache metrics
    cacheOperationsTotal *prometheus.CounterVec
    cacheHitRatio       *prometheus.GaugeVec
    
    // System metrics
    memoryUsage         prometheus.Gauge
    cpuUsage           prometheus.Gauge
    goroutinesActive   prometheus.Gauge
}

func NewMetricsService(namespace string) *MetricsService {
    return &MetricsService{
        httpRequestsTotal: promauto.NewCounterVec(
            prometheus.CounterOpts{
                Namespace: namespace,
                Name:      "http_requests_total",
                Help:      "Total number of HTTP requests",
            },
            []string{"method", "endpoint", "status_code"},
        ),
        
        httpRequestDuration: promauto.NewHistogramVec(
            prometheus.HistogramOpts{
                Namespace: namespace,
                Name:      "http_request_duration_seconds",
                Help:      "HTTP request duration in seconds",
                Buckets:   prometheus.DefBuckets,
            },
            []string{"method", "endpoint"},
        ),
        
        businessOperationsTotal: promauto.NewCounterVec(
            prometheus.CounterOpts{
                Namespace: namespace,
                Name:      "business_operations_total",
                Help:      "Total number of business operations",
            },
            []string{"domain", "operation", "status"},
        ),
        
        businessOperationDuration: promauto.NewHistogramVec(
            prometheus.HistogramOpts{
                Namespace: namespace,
                Name:      "business_operation_duration_seconds",
                Help:      "Business operation duration in seconds",
                Buckets:   []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 2, 5},
            },
            []string{"domain", "operation"},
        ),
        
        dbQueryDuration: promauto.NewHistogramVec(
            prometheus.HistogramOpts{
                Namespace: namespace,
                Name:      "database_query_duration_seconds",
                Help:      "Database query duration in seconds",
                Buckets:   []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1},
            },
            []string{"operation", "table"},
        ),
        
        cacheOperationsTotal: promauto.NewCounterVec(
            prometheus.CounterOpts{
                Namespace: namespace,
                Name:      "cache_operations_total",
                Help:      "Total number of cache operations",
            },
            []string{"operation", "result"},
        ),
    }
}

// HTTP Metrics Middleware
func (m *MetricsService) HTTPMetricsMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        
        c.Next()
        
        duration := time.Since(start).Seconds()
        statusCode := strconv.Itoa(c.Writer.Status())
        
        m.httpRequestsTotal.WithLabelValues(
            c.Request.Method,
            c.FullPath(),
            statusCode,
        ).Inc()
        
        m.httpRequestDuration.WithLabelValues(
            c.Request.Method,
            c.FullPath(),
        ).Observe(duration)
    }
}

// Business Operation Metrics
func (m *MetricsService) RecordBusinessOperation(domain, operation string, duration time.Duration, success bool) {
    status := "success"
    if !success {
        status = "error"
    }
    
    m.businessOperationsTotal.WithLabelValues(domain, operation, status).Inc()
    m.businessOperationDuration.WithLabelValues(domain, operation).Observe(duration.Seconds())
}

// Database Metrics
func (m *MetricsService) RecordDBQuery(operation, table string, duration time.Duration) {
    m.dbQueryDuration.WithLabelValues(operation, table).Observe(duration.Seconds())
}

// Cache Metrics
func (m *MetricsService) RecordCacheOperation(operation string, hit bool) {
    result := "miss"
    if hit {
        result = "hit"
    }
    m.cacheOperationsTotal.WithLabelValues(operation, result).Inc()
}

// Custom Business Metrics
type OrderMetrics struct {
    ordersCreated     prometheus.Counter
    ordersCompleted   prometheus.Counter
    orderValue        prometheus.Histogram
    orderProcessingTime prometheus.Histogram
}

func NewOrderMetrics() *OrderMetrics {
    return &OrderMetrics{
        ordersCreated: promauto.NewCounter(prometheus.CounterOpts{
            Name: "orders_created_total",
            Help: "Total number of orders created",
        }),
        
        ordersCompleted: promauto.NewCounter(prometheus.CounterOpts{
            Name: "orders_completed_total",
            Help: "Total number of orders completed",
        }),
        
        orderValue: promauto.NewHistogram(prometheus.HistogramOpts{
            Name:    "order_value_dollars",
            Help:    "Order value in dollars",
            Buckets: []float64{10, 50, 100, 500, 1000, 5000},
        }),
        
        orderProcessingTime: promauto.NewHistogram(prometheus.HistogramOpts{
            Name:    "order_processing_duration_seconds",
            Help:    "Time to process an order",
            Buckets: []float64{1, 5, 10, 30, 60, 300, 600},
        }),
    }
}
```

### **3.3 Health Checks & SLA Monitoring**
```go
// src/common/health/health_checker.go
package health

import (
    "context"
    "sync"
    "time"
)

type HealthStatus string

const (
    StatusHealthy   HealthStatus = "healthy"
    StatusUnhealthy HealthStatus = "unhealthy"
    StatusDegraded  HealthStatus = "degraded"
)

type HealthCheck interface {
    Name() string
    Check(ctx context.Context) HealthCheckResult
}

type HealthCheckResult struct {
    Status    HealthStatus          `json:"status"`
    Message   string               `json:"message,omitempty"`
    Details   map[string]interface{} `json:"details,omitempty"`
    Duration  time.Duration        `json:"duration"`
    Timestamp time.Time            `json:"timestamp"`
}

type HealthChecker struct {
    checks map[string]HealthCheck
    cache  map[string]HealthCheckResult
    mutex  sync.RWMutex
    logger logger.Logger
}

func NewHealthChecker(logger logger.Logger) *HealthChecker {
    return &HealthChecker{
        checks: make(map[string]HealthCheck),
        cache:  make(map[string]HealthCheckResult),
        logger: logger,
    }
}

func (hc *HealthChecker) AddCheck(check HealthCheck) {
    hc.mutex.Lock()
    defer hc.mutex.Unlock()
    
    hc.checks[check.Name()] = check
}

func (hc *HealthChecker) CheckAll(ctx context.Context) map[string]HealthCheckResult {
    hc.mutex.RLock()
    checks := make(map[string]HealthCheck)
    for name, check := range hc.checks {
        checks[name] = check
    }
    hc.mutex.RUnlock()
    
    results := make(map[string]HealthCheckResult)
    var wg sync.WaitGroup
    var mu sync.Mutex
    
    for name, check := range checks {
        wg.Add(1)
        go func(name string, check HealthCheck) {
            defer wg.Done()
            
            result := check.Check(ctx)
            
            mu.Lock()
            results[name] = result
            mu.Unlock()
        }(name, check)
    }
    
    wg.Wait()
    
    // Update cache
    hc.mutex.Lock()
    for name, result := range results {
        hc.cache[name] = result
    }
    hc.mutex.Unlock()
    
    return results
}

// Database Health Check
type DatabaseHealthCheck struct {
    name string
    db   database.DB
}

func NewDatabaseHealthCheck(name string, db database.DB) *DatabaseHealthCheck {
    return &DatabaseHealthCheck{name: name, db: db}
}

func (dhc *DatabaseHealthCheck) Name() string {
    return dhc.name
}

func (dhc *DatabaseHealthCheck) Check(ctx context.Context) HealthCheckResult {
    start := time.Now()
    
    err := dhc.db.Ping(ctx)
    duration := time.Since(start)
    
    if err != nil {
        return HealthCheckResult{
            Status:    StatusUnhealthy,
            Message:   err.Error(),
            Duration:  duration,
            Timestamp: time.Now(),
        }
    }
    
    // Check connection pool stats
    stats := dhc.db.Stats()
    details := map[string]interface{}{
        "open_connections":     stats.OpenConnections,
        "in_use_connections":   stats.InUse,
        "idle_connections":     stats.Idle,
        "max_open_connections": stats.MaxOpenConnections,
    }
    
    status := StatusHealthy
    if stats.OpenConnections > int(float64(stats.MaxOpenConnections)*0.8) {
        status = StatusDegraded
    }
    
    return HealthCheckResult{
        Status:    status,
        Details:   details,
        Duration:  duration,
        Timestamp: time.Now(),
    }
}

// Redis Health Check
type RedisHealthCheck struct {
    name   string
    client redis.UniversalClient
}

func (rhc *RedisHealthCheck) Check(ctx context.Context) HealthCheckResult {
    start := time.Now()
    
    err := rhc.client.Ping(ctx).Err()
    duration := time.Since(start)
    
    if err != nil {
        return HealthCheckResult{
            Status:    StatusUnhealthy,
            Message:   err.Error(),
            Duration:  duration,
            Timestamp: time.Now(),
        }
    }
    
    // Get Redis info
    info, err := rhc.client.Info(ctx).Result()
    if err != nil {
        return HealthCheckResult{
            Status:    StatusDegraded,
            Message:   "Could not get Redis info",
            Duration:  duration,
            Timestamp: time.Now(),
        }
    }
    
    details := map[string]interface{}{
        "redis_version": parseRedisInfo(info, "redis_version"),
        "used_memory":   parseRedisInfo(info, "used_memory"),
        "connected_clients": parseRedisInfo(info, "connected_clients"),
    }
    
    return HealthCheckResult{
        Status:    StatusHealthy,
        Details:   details,
        Duration:  duration,
        Timestamp: time.Now(),
    }
}

// SLA Monitor
type SLAMonitor struct {
    target           float64 // Target availability (e.g., 0.999 for 99.9%)
    window           time.Duration
    measurements     []SLAMeasurement
    mutex           sync.RWMutex
    alertThreshold  float64
    alertCallback   func(sla float64, target float64)
}

type SLAMeasurement struct {
    Timestamp time.Time
    Success   bool
    Duration  time.Duration
}

func NewSLAMonitor(target float64, window time.Duration) *SLAMonitor {
    return &SLAMonitor{
        target:         target,
        window:         window,
        measurements:   make([]SLAMeasurement, 0),
        alertThreshold: target - 0.01, // Alert if SLA drops 1% below target
    }
}

func (sla *SLAMonitor) RecordMeasurement(success bool, duration time.Duration) {
    sla.mutex.Lock()
    defer sla.mutex.Unlock()
    
    measurement := SLAMeasurement{
        Timestamp: time.Now(),
        Success:   success,
        Duration:  duration,
    }
    
    sla.measurements = append(sla.measurements, measurement)
    
    // Clean old measurements
    cutoff := time.Now().Add(-sla.window)
    var validMeasurements []SLAMeasurement
    for _, m := range sla.measurements {
        if m.Timestamp.After(cutoff) {
            validMeasurements = append(validMeasurements, m)
        }
    }
    sla.measurements = validMeasurements
    
    // Check if we need to alert
    if len(sla.measurements) > 10 { // Only alert if we have enough data
        currentSLA := sla.getCurrentSLA()
        if currentSLA < sla.alertThreshold && sla.alertCallback != nil {
            sla.alertCallback(currentSLA, sla.target)
        }
    }
}

func (sla *SLAMonitor) GetCurrentSLA() float64 {
    sla.mutex.RLock()
    defer sla.mutex.RUnlock()
    
    return sla.getCurrentSLA()
}

func (sla *SLAMonitor) getCurrentSLA() float64 {
    if len(sla.measurements) == 0 {
        return 1.0
    }
    
    successful := 0
    for _, m := range sla.measurements {
        if m.Success {
            successful++
        }
    }
    
    return float64(successful) / float64(len(sla.measurements))
}
```

---

## 🔐 **4. ADVANCED SECURITY PATTERNS**

### **4.1 Multi-Factor Authentication (MFA)**
```go
// src/common/security/mfa.go
package security

import (
    "crypto/rand"
    "encoding/base32"
    "fmt"
    "time"
    
    "github.com/pquerna/otp"
    "github.com/pquerna/otp/totp"
)

type MFAService struct {
    issuer string
    cache  cache.Cache
    sms    SMSService
    email  EmailService
}

type MFAMethod string

const (
    MFAMethodTOTP  MFAMethod = "totp"
    MFAMethodSMS   MFAMethod = "sms"
    MFAMethodEmail MFAMethod = "email"
)

type MFASetupResult struct {
    Secret    string `json:"secret"`
    QRCodeURL string `json:"qr_code_url"`
    BackupCodes []string `json:"backup_codes"`
}

func (mfa *MFAService) SetupTOTP(userID, email string) (*MFASetupResult, error) {
    key, err := totp.Generate(totp.GenerateOpts{
        Issuer:      mfa.issuer,
        AccountName: email,
        SecretSize:  32,
    })
    if err != nil {
        return nil, err
    }
    
    // Generate backup codes
    backupCodes := make([]string, 10)
    for i := range backupCodes {
        code := make([]byte, 8)
        rand.Read(code)
        backupCodes[i] = base32.StdEncoding.EncodeToString(code)[:8]
    }
    
    return &MFASetupResult{
        Secret:      key.Secret(),
        QRCodeURL:   key.URL(),
        BackupCodes: backupCodes,
    }, nil
}

func (mfa *MFAService) VerifyTOTP(secret, code string) bool {
    return totp.Validate(code, secret)
}

func (mfa *MFAService) SendSMSCode(userID, phoneNumber string) error {
    code := generateRandomCode(6)
    
    // Store code in cache with 5 minute expiry
    key := fmt.Sprintf("mfa:sms:%s", userID)
    mfa.cache.Set(context.Background(), key, code, 5*time.Minute)
    
    // Send SMS
    message := fmt.Sprintf("Your verification code is: %s", code)
    return mfa.sms.SendSMS(phoneNumber, message)
}

func (mfa *MFAService) VerifySMSCode(userID, code string) bool {
    key := fmt.Sprintf("mfa:sms:%s", userID)
    storedCode, err := mfa.cache.Get(context.Background(), key)
    if err != nil {
        return false
    }
    
    if storedCode.(string) == code {
        // Delete code after successful verification
        mfa.cache.Delete(context.Background(), key)
        return true
    }
    
    return false
}

// Risk-based Authentication
type RiskAssessment struct {
    Score      float64           `json:"score"`
    Factors    map[string]float64 `json:"factors"`
    Requires2FA bool             `json:"requires_2fa"`
    Action     string            `json:"action"`
}

type RiskAssessor struct {
    geoIP      GeoIPService
    deviceRepo DeviceRepository
    loginRepo  LoginHistoryRepository
}

func (ra *RiskAssessor) AssessRisk(ctx context.Context, userID, ip, userAgent string) (*RiskAssessment, error) {
    factors := make(map[string]float64)
    
    // Geographic risk
    geoRisk, err := ra.assessGeographicRisk(ctx, userID, ip)
    if err == nil {
        factors["geographic"] = geoRisk
    }
    
    // Device risk
    deviceRisk, err := ra.assessDeviceRisk(ctx, userID, userAgent)
    if err == nil {
        factors["device"] = deviceRisk
    }
    
    // Time-based risk
    timeRisk := ra.assessTimeRisk(ctx, userID)
    factors["time"] = timeRisk
    
    // Behavioral risk
    behaviorRisk, err := ra.assessBehavioralRisk(ctx, userID)
    if err == nil {
        factors["behavior"] = behaviorRisk
    }
    
    // Calculate overall risk score
    totalScore := 0.0
    for _, score := range factors {
        totalScore += score
    }
    avgScore := totalScore / float64(len(factors))
    
    // Determine action based on risk score
    var action string
    requires2FA := false
    
    switch {
    case avgScore < 0.3:
        action = "allow"
    case avgScore < 0.7:
        action = "require_2fa"
        requires2FA = true
    default:
        action = "block"
    }
    
    return &RiskAssessment{
        Score:      avgScore,
        Factors:    factors,
        Requires2FA: requires2FA,
        Action:     action,
    }, nil
}

func (ra *RiskAssessor) assessGeographicRisk(ctx context.Context, userID, ip string) (float64, error) {
    // Get user's typical locations
    locations, err := ra.loginRepo.GetUserLocations(ctx, userID, 30) // Last 30 days
    if err != nil {
        return 0.5, err // Default medium risk
    }
    
    // Get current location
    currentLocation, err := ra.geoIP.GetLocation(ip)
    if err != nil {
        return 0.5, err
    }
    
    // Check if current location is familiar
    for _, loc := range locations {
        if distance(currentLocation, loc) < 100 { // Within 100km
            return 0.1, nil // Low risk
        }
    }
    
    // Check if it's in the same country
    for _, loc := range locations {
        if currentLocation.Country == loc.Country {
            return 0.4, nil // Medium risk
        }
    }
    
    return 0.8, nil // High risk - new country
}
```

### **4.2 Advanced Authorization (ABAC - Attribute-Based Access Control)**
```go
// src/common/security/abac.go
package security

import (
    "context"
    "encoding/json"
    "fmt"
)

type Attribute struct {
    Name  string      `json:"name"`
    Value interface{} `json:"value"`
}

type Subject struct {
    ID         string               `json:"id"`
    Type       string               `json:"type"`
    Attributes map[string]interface{} `json:"attributes"`
}

type Resource struct {
    ID         string               `json:"id"`
    Type       string               `json:"type"`
    Attributes map[string]interface{} `json:"attributes"`
}

type Action struct {
    Name       string               `json:"name"`
    Attributes map[string]interface{} `json:"attributes"`
}

type Environment struct {
    Attributes map[string]interface{} `json:"attributes"`
}

type PolicyRule struct {
    ID          string                 `json:"id"`
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    Target      PolicyTarget           `json:"target"`
    Condition   string                 `json:"condition"` // Expression language
    Effect      PolicyEffect           `json:"effect"`
    Priority    int                    `json:"priority"`
}

type PolicyTarget struct {
    Subjects   []string `json:"subjects"`   // Subject types
    Resources  []string `json:"resources"`  // Resource types
    Actions    []string `json:"actions"`    // Action names
}

type PolicyEffect string

const (
    EffectAllow PolicyEffect = "allow"
    EffectDeny  PolicyEffect = "deny"
)

type AuthorizationRequest struct {
    Subject     Subject     `json:"subject"`
    Resource    Resource    `json:"resource"`
    Action      Action      `json:"action"`
    Environment Environment `json:"environment"`
}

type AuthorizationResponse struct {
    Decision    PolicyEffect          `json:"decision"`
    Reason      string               `json:"reason"`
    AppliedRules []string            `json:"applied_rules"`
    Obligations []string             `json:"obligations"`
}

type ABACEngine struct {
    policies       []PolicyRule
    ruleEvaluator  RuleEvaluator
    logger         logger.Logger
}

func NewABACEngine(policies []PolicyRule, logger logger.Logger) *ABACEngine {
    return &ABACEngine{
        policies:      policies,
        ruleEvaluator: NewRuleEvaluator(),
        logger:        logger,
    }
}

func (engine *ABACEngine) Authorize(ctx context.Context, req AuthorizationRequest) (*AuthorizationResponse, error) {
    engine.logger.Debug(ctx, "Evaluating authorization request",
        logger.String("subject_id", req.Subject.ID),
        logger.String("resource_id", req.Resource.ID),
        logger.String("action", req.Action.Name))
    
    var applicableRules []PolicyRule
    var appliedRules []string
    decision := EffectDeny // Default deny
    reason := "No applicable rules found"
    
    // Find applicable rules
    for _, rule := range engine.policies {
        if engine.isRuleApplicable(rule, req) {
            applicableRules = append(applicableRules, rule)
        }
    }
    
    // Sort by priority (higher priority first)
    sort.Slice(applicableRules, func(i, j int) bool {
        return applicableRules[i].Priority > applicableRules[j].Priority
    })
    
    // Evaluate rules
    for _, rule := range applicableRules {
        matches, err := engine.evaluateRule(ctx, rule, req)
        if err != nil {
            engine.logger.Error(ctx, err, "Error evaluating rule", 
                logger.String("rule_id", rule.ID))
            continue
        }
        
        if matches {
            appliedRules = append(appliedRules, rule.ID)
            decision = rule.Effect
            reason = fmt.Sprintf("Rule '%s' applied", rule.Name)
            
            // If we have an explicit deny, stop processing
            if rule.Effect == EffectDeny {
                break
            }
        }
    }
    
    response := &AuthorizationResponse{
        Decision:     decision,
        Reason:       reason,
        AppliedRules: appliedRules,
    }
    
    engine.logger.Info(ctx, "Authorization decision made",
        logger.String("decision", string(decision)),
        logger.String("reason", reason),
        logger.String("applied_rules", strings.Join(appliedRules, ",")))
    
    return response, nil
}

func (engine *ABACEngine) isRuleApplicable(rule PolicyRule, req AuthorizationRequest) bool {
    // Check if subject type matches
    if len(rule.Target.Subjects) > 0 {
        found := false
        for _, subjectType := range rule.Target.Subjects {
            if subjectType == req.Subject.Type {
                found = true
                break
            }
        }
        if !found {
            return false
        }
    }
    
    // Check if resource type matches
    if len(rule.Target.Resources) > 0 {
        found := false
        for _, resourceType := range rule.Target.Resources {
            if resourceType == req.Resource.Type {
                found = true
                break
            }
        }
        if !found {
            return false
        }
    }
    
    // Check if action matches
    if len(rule.Target.Actions) > 0 {
        found := false
        for _, action := range rule.Target.Actions {
            if action == req.Action.Name {
                found = true
                break
            }
        }
        if !found {
            return false
        }
    }
    
    return true
}

// Rule Evaluator using expression language
type RuleEvaluator struct {
    // Could use libraries like govaluate or expr
}

func (re *RuleEvaluator) Evaluate(condition string, context map[string]interface{}) (bool, error) {
    // Example conditions:
    // "subject.department == 'finance' && resource.classification == 'confidential'"
    // "subject.role in ['admin', 'manager'] && environment.time_of_day between '09:00' and '17:00'"
    // "resource.owner == subject.id || subject.permissions contains 'admin'"
    
    // This would typically use an expression evaluation library
    // For now, simplified implementation
    
    switch condition {
    case "subject.role == 'admin'":
        if role, ok := context["subject.role"]; ok {
            return role == "admin", nil
        }
    case "resource.owner == subject.id":
        owner, ownerOk := context["resource.owner"]
        subjectID, subjectOk := context["subject.id"]
        if ownerOk && subjectOk {
            return owner == subjectID, nil
        }
    }
    
    return false, fmt.Errorf("unsupported condition: %s", condition)
}

// Example policies
var SamplePolicies = []PolicyRule{
    {
        ID:          "admin-full-access",
        Name:        "Admin Full Access",
        Description: "Administrators have full access to all resources",
        Target: PolicyTarget{
            Subjects:  []string{"user"},
            Resources: []string{"*"},
            Actions:   []string{"*"},
        },
        Condition: "subject.role == 'admin'",
        Effect:    EffectAllow,
        Priority:  100,
    },
    {
        ID:          "owner-access",
        Name:        "Resource Owner Access",
        Description: "Users can access resources they own",
        Target: PolicyTarget{
            Subjects:  []string{"user"},
            Resources: []string{"*"},
            Actions:   []string{"read", "update", "delete"},
        },
        Condition: "resource.owner == subject.id",
        Effect:    EffectAllow,
        Priority:  50,
    },
    {
        ID:          "business-hours-only",
        Name:        "Business Hours Only",
        Description: "Sensitive operations only during business hours",
        Target: PolicyTarget{
            Resources: []string{"financial-data"},
            Actions:   []string{"create", "update", "delete"},
        },
        Condition: "environment.time_of_day between '09:00' and '17:00'",
        Effect:    EffectDeny,
        Priority:  75,
    },
}
```

---

## 🚀 **5. PERFORMANCE OPTIMIZATION PATTERNS**

### **5.1 Advanced Caching Strategies**
```go
// src/common/cache/advanced_cache.go
package cache

import (
    "context"
    "encoding/json"
    "fmt"
    "sync"
    "time"
)

// Multi-level Cache
type MultiLevelCache struct {
    l1Cache Cache // In-memory cache (fast)
    l2Cache Cache // Distributed cache (Redis)
    l3Cache Cache // Persistent cache (Database)
    
    l1TTL time.Duration
    l2TTL time.Duration
    l3TTL time.Duration
    
    metrics *CacheMetrics
}

func NewMultiLevelCache(l1, l2, l3 Cache, l1TTL, l2TTL, l3TTL time.Duration) *MultiLevelCache {
    return &MultiLevelCache{
        l1Cache: l1,
        l2Cache: l2,
        l3Cache: l3,
        l1TTL:   l1TTL,
        l2TTL:   l2TTL,
        l3TTL:   l3TTL,
        metrics: NewCacheMetrics(),
    }
}

func (mlc *MultiLevelCache) Get(ctx context.Context, key string) (interface{}, error) {
    // Try L1 cache first
    if value, err := mlc.l1Cache.Get(ctx, key); err == nil {
        mlc.metrics.RecordHit("l1")
        return value, nil
    }
    mlc.metrics.RecordMiss("l1")
    
    // Try L2 cache
    if value, err := mlc.l2Cache.Get(ctx, key); err == nil {
        mlc.metrics.RecordHit("l2")
        
        // Populate L1 cache
        go mlc.l1Cache.Set(ctx, key, value, mlc.l1TTL)
        
        return value, nil
    }
    mlc.metrics.RecordMiss("l2")
    
    // Try L3 cache
    if value, err := mlc.l3Cache.Get(ctx, key); err == nil {
        mlc.metrics.RecordHit("l3")
        
        // Populate L2 and L1 caches
        go func() {
            mlc.l2Cache.Set(ctx, key, value, mlc.l2TTL)
            mlc.l1Cache.Set(ctx, key, value, mlc.l1TTL)
        }()
        
        return value, nil
    }
    mlc.metrics.RecordMiss("l3")
    
    return nil, ErrCacheMiss
}

func (mlc *MultiLevelCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
    // Set in all cache levels
    var wg sync.WaitGroup
    errors := make([]error, 3)
    
    wg.Add(3)
    
    // L1 cache
    go func() {
        defer wg.Done()
        errors[0] = mlc.l1Cache.Set(ctx, key, value, mlc.l1TTL)
    }()
    
    // L2 cache
    go func() {
        defer wg.Done()
        errors[1] = mlc.l2Cache.Set(ctx, key, value, mlc.l2TTL)
    }()
    
    // L3 cache
    go func() {
        defer wg.Done()
        errors[2] = mlc.l3Cache.Set(ctx, key, value, mlc.l3TTL)
    }()
    
    wg.Wait()
    
    // Return first error encountered
    for _, err := range errors {
        if err != nil {
            return err
        }
    }
    
    return nil
}

// Cache-Aside Pattern with Refresh-Ahead
type RefreshAheadCache struct {
    cache       Cache
    dataLoader  DataLoader
    refreshPool *sync.Pool
    
    refreshThreshold float64 // Refresh when TTL is 20% remaining
    refreshInProgress sync.Map
}

type DataLoader interface {
    Load(ctx context.Context, key string) (interface{}, error)
}

func NewRefreshAheadCache(cache Cache, loader DataLoader) *RefreshAheadCache {
    return &RefreshAheadCache{
        cache:            cache,
        dataLoader:       loader,
        refreshThreshold: 0.2,
        refreshPool: &sync.Pool{
            New: func() interface{} {
                return make(chan struct{}, 1)
            },
        },
    }
}

func (rac *RefreshAheadCache) Get(ctx context.Context, key string) (interface{}, error) {
    // Try to get from cache
    value, err := rac.cache.Get(ctx, key)
    if err == nil {
        // Check if we should refresh
        if rac.shouldRefresh(ctx, key) {
            go rac.refreshAsync(ctx, key)
        }
        return value, nil
    }
    
    // Cache miss - load synchronously
    return rac.loadAndCache(ctx, key)
}

func (rac *RefreshAheadCache) shouldRefresh(ctx context.Context, key string) bool {
    ttl, err := rac.cache.TTL(ctx, key)
    if err != nil {
        return false
    }
    
    // Refresh if TTL is below threshold (assuming 1 hour original TTL)
    originalTTL := 1 * time.Hour
    remaining := ttl.Seconds() / originalTTL.Seconds()
    
    return remaining < rac.refreshThreshold
}

func (rac *RefreshAheadCache) refreshAsync(ctx context.Context, key string) {
    // Prevent multiple concurrent refreshes for the same key
    if _, loaded := rac.refreshInProgress.LoadOrStore(key, true); loaded {
        return
    }
    defer rac.refreshInProgress.Delete(key)
    
    // Load fresh data
    value, err := rac.dataLoader.Load(ctx, key)
    if err != nil {
        return // Keep old data in cache
    }
    
    // Update cache
    rac.cache.Set(ctx, key, value, 1*time.Hour)
}

func (rac *RefreshAheadCache) loadAndCache(ctx context.Context, key string) (interface{}, error) {
    value, err := rac.dataLoader.Load(ctx, key)
    if err != nil {
        return nil, err
    }
    
    // Cache the loaded value
    go rac.cache.Set(ctx, key, value, 1*time.Hour)
    
    return value, nil
}

// Write-Behind Cache
type WriteBehindCache struct {
    cache       Cache
    dataStore   DataStore
    writeQueue  chan WriteOperation
    batchSize   int
    flushInterval time.Duration
}

type WriteOperation struct {
    Key   string
    Value interface{}
    Op    string // "set" or "delete"
}

type DataStore interface {
    BatchWrite(ctx context.Context, operations []WriteOperation) error
}

func NewWriteBehindCache(cache Cache, store DataStore) *WriteBehindCache {
    wbc := &WriteBehindCache{
        cache:         cache,
        dataStore:     store,
        writeQueue:    make(chan WriteOperation, 1000),
        batchSize:     100,
        flushInterval: 5 * time.Second,
    }
    
    // Start background writer
    go wbc.backgroundWriter()
    
    return wbc
}

func (wbc *WriteBehindCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
    // Write to cache immediately
    if err := wbc.cache.Set(ctx, key, value, ttl); err != nil {
        return err
    }
    
    // Queue for background write to persistent store
    select {
    case wbc.writeQueue <- WriteOperation{Key: key, Value: value, Op: "set"}:
        return nil
    default:
        // Queue full, write synchronously
        return wbc.dataStore.BatchWrite(ctx, []WriteOperation{{Key: key, Value: value, Op: "set"}})
    }
}

func (wbc *WriteBehindCache) backgroundWriter() {
    ticker := time.NewTicker(wbc.flushInterval)
    defer ticker.Stop()
    
    operations := make([]WriteOperation, 0, wbc.batchSize)
    
    for {
        select {
        case op := <-wbc.writeQueue:
            operations = append(operations, op)
            
            if len(operations) >= wbc.batchSize {
                wbc.flushOperations(operations)
                operations = operations[:0]
            }
            
        case <-ticker.C:
            if len(operations) > 0 {
                wbc.flushOperations(operations)
                operations = operations[:0]
            }
        }
    }
}

func (wbc *WriteBehindCache) flushOperations(operations []WriteOperation) {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    if err := wbc.dataStore.BatchWrite(ctx, operations); err != nil {
        // Handle error - could retry, log, or alert
        logger.Error(ctx, err, "Failed to flush cache operations to data store")
    }
}
```

### **5.2 Database Optimization Patterns**
```go
// src/common/database/optimization.go
package database

import (
    "context"
    "database/sql"
    "sync"
    "time"
)

// Connection Pool Manager
type PoolManager struct {
    pools   map[string]*sql.DB
    mutex   sync.RWMutex
    metrics *PoolMetrics
}

type PoolConfig struct {
    MaxOpenConns    int
    MaxIdleConns    int
    ConnMaxLifetime time.Duration
    ConnMaxIdleTime time.Duration
}

func NewPoolManager() *PoolManager {
    return &PoolManager{
        pools:   make(map[string]*sql.DB),
        metrics: NewPoolMetrics(),
    }
}

func (pm *PoolManager) GetPool(name string) *sql.DB {
    pm.mutex.RLock()
    defer pm.mutex.RUnlock()
    
    return pm.pools[name]
}

func (pm *PoolManager) CreatePool(name, dsn string, config PoolConfig) error {
    pm.mutex.Lock()
    defer pm.mutex.Unlock()
    
    db, err := sql.Open("postgres", dsn)
    if err != nil {
        return err
    }
    
    db.SetMaxOpenConns(config.MaxOpenConns)
    db.SetMaxIdleConns(config.MaxIdleConns)
    db.SetConnMaxLifetime(config.ConnMaxLifetime)
    db.SetConnMaxIdleTime(config.ConnMaxIdleTime)
    
    pm.pools[name] = db
    
    // Start monitoring
    go pm.monitorPool(name, db)
    
    return nil
}

func (pm *PoolManager) monitorPool(name string, db *sql.DB) {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for range ticker.C {
        stats := db.Stats()
        pm.metrics.RecordPoolStats(name, stats)
    }
}

// Read/Write Splitting
type ReadWriteDB struct {
    writeDB *sql.DB
    readDBs []*sql.DB
    
    readIndex int64
    mutex     sync.RWMutex
}

func NewReadWriteDB(writeDB *sql.DB, readDBs []*sql.DB) *ReadWriteDB {
    return &ReadWriteDB{
        writeDB: writeDB,
        readDBs: readDBs,
    }
}

func (rw *ReadWriteDB) Write() *sql.DB {
    return rw.writeDB
}

func (rw *ReadWriteDB) Read() *sql.DB {
    if len(rw.readDBs) == 0 {
        return rw.writeDB
    }
    
    index := atomic.AddInt64(&rw.readIndex, 1) % int64(len(rw.readDBs))
    return rw.readDBs[index]
}

// Query Builder with Optimization
type QueryBuilder struct {
    query      strings.Builder
    args       []interface{}
    hints      []string
    indexes    []string
    explain    bool
}

func NewQueryBuilder() *QueryBuilder {
    return &QueryBuilder{
        args: make([]interface{}, 0),
    }
}

func (qb *QueryBuilder) Select(columns ...string) *QueryBuilder {
    qb.query.WriteString("SELECT ")
    qb.query.WriteString(strings.Join(columns, ", "))
    return qb
}

func (qb *QueryBuilder) From(table string) *QueryBuilder {
    qb.query.WriteString(" FROM ")
    qb.query.WriteString(table)
    return qb
}

func (qb *QueryBuilder) UseIndex(index string) *QueryBuilder {
    qb.indexes = append(qb.indexes, index)
    return qb
}

func (qb *QueryBuilder) Hint(hint string) *QueryBuilder {
    qb.hints = append(qb.hints, hint)
    return qb
}

func (qb *QueryBuilder) Where(condition string, args ...interface{}) *QueryBuilder {
    qb.query.WriteString(" WHERE ")
    qb.query.WriteString(condition)
    qb.args = append(qb.args, args...)
    return qb
}

func (qb *QueryBuilder) OrderBy(column, direction string) *QueryBuilder {
    qb.query.WriteString(" ORDER BY ")
    qb.query.WriteString(column)
    qb.query.WriteString(" ")
    qb.query.WriteString(direction)
    return qb
}

func (qb *QueryBuilder) Limit(limit int) *QueryBuilder {
    qb.query.WriteString(" LIMIT ")
    qb.query.WriteString(strconv.Itoa(limit))
    return qb
}

func (qb *QueryBuilder) Build() (string, []interface{}) {
    query := qb.query.String()
    
    // Add hints if any
    if len(qb.hints) > 0 {
        hintStr := "/*+ " + strings.Join(qb.hints, ", ") + " */"
        query = hintStr + " " + query
    }
    
    // Add index hints if any (PostgreSQL specific)
    if len(qb.indexes) > 0 {
        // This would be database-specific implementation
    }
    
    return query, qb.args
}

// Batch Operations
type BatchProcessor struct {
    db        *sql.DB
    batchSize int
}

func NewBatchProcessor(db *sql.DB, batchSize int) *BatchProcessor {
    return &BatchProcessor{
        db:        db,
        batchSize: batchSize,
    }
}

func (bp *BatchProcessor) BatchInsert(ctx context.Context, table string, columns []string, values [][]interface{}) error {
    if len(values) == 0 {
        return nil
    }
    
    // Process in batches
    for i := 0; i < len(values); i += bp.batchSize {
        end := i + bp.batchSize
        if end > len(values) {
            end = len(values)
        }
        
        batch := values[i:end]
        if err := bp.insertBatch(ctx, table, columns, batch); err != nil {
            return err
        }
    }
    
    return nil
}

func (bp *BatchProcessor) insertBatch(ctx context.Context, table string, columns []string, values [][]interface{}) error {
    if len(values) == 0 {
        return nil
    }
    
    // Build bulk insert query
    placeholders := make([]string, len(values))
    args := make([]interface{}, 0, len(values)*len(columns))
    
    for i, row := range values {
        rowPlaceholders := make([]string, len(columns))
        for j := range columns {
            argIndex := i*len(columns) + j + 1
            rowPlaceholders[j] = fmt.Sprintf("$%d", argIndex)
            args = append(args, row[j])
        }
        placeholders[i] = "(" + strings.Join(rowPlaceholders, ", ") + ")"
    }
    
    query := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s",
        table,
        strings.Join(columns, ", "),
        strings.Join(placeholders, ", "))
    
    _, err := bp.db.ExecContext(ctx, query, args...)
    return err
}

// Prepared Statement Pool
type PreparedStatementPool struct {
    statements sync.Map // map[string]*sql.Stmt
    db         *sql.DB
}

func NewPreparedStatementPool(db *sql.DB) *PreparedStatementPool {
    return &PreparedStatementPool{
        db: db,
    }
}

func (psp *PreparedStatementPool) GetStatement(query string) (*sql.Stmt, error) {
    if stmt, ok := psp.statements.Load(query); ok {
        return stmt.(*sql.Stmt), nil
    }
    
    stmt, err := psp.db.Prepare(query)
    if err != nil {
        return nil, err
    }
    
    psp.statements.Store(query, stmt)
    return stmt, nil
}

func (psp *PreparedStatementPool) Close() error {
    var lastErr error
    psp.statements.Range(func(key, value interface{}) bool {
        if stmt, ok := value.(*sql.Stmt); ok {
            if err := stmt.Close(); err != nil {
                lastErr = err
            }
        }
        return true
    })
    return lastErr
}
```

---

Tôi đã mở rộng đáng kể khuôn mẫu backend với các patterns và techniques nâng cao:

## 🎯 **CÁC PATTERNS NÂNG CAO ĐÃ BỔ SUNG**

### **1. Event-Driven Architecture**
- ✅ Event Sourcing với Aggregate Pattern
- ✅ CQRS (Command Query Responsibility Segregation) 
- ✅ Saga Pattern cho distributed transactions
- ✅ Event streaming và replay capabilities

### **2. Microservices Patterns**
- ✅ Service Discovery & Registry (Consul-based)
- ✅ Circuit Breaker với health monitoring
- ✅ API Gateway với routing, load balancing
- ✅ Rate limiting và request throttling

### **3. Advanced Monitoring**
- ✅ Distributed tracing với custom spans
- ✅ Business metrics và KPI tracking
- ✅ SLA monitoring với alerting
- ✅ Health checks multi-level

### **4. Enterprise Security**
- ✅ Multi-Factor Authentication (TOTP, SMS, Email)
- ✅ Risk-based authentication
- ✅ ABAC (Attribute-Based Access Control)
- ✅ Advanced policy engine

### **5. Performance Optimization**
- ✅ Multi-level caching (L1, L2, L3)
- ✅ Cache-aside với refresh-ahead
- ✅ Write-behind caching pattern
- ✅ Database optimization (connection pooling, read/write splitting)

## 🚀 **GIÁ TRỊ NÂNG CAP**

Với các patterns nâng cao này, khuôn mẫu giờ đây có thể:

### **Enterprise-Ready Features**
- 🏢 **Scalable Architecture** - Microservices, event-driven
- 🔒 **Enterprise Security** - ABAC, MFA, risk assessment  
- 📊 **Advanced Monitoring** - SLA tracking, business metrics
- ⚡ **High Performance** - Multi-level caching, database optimization

### **Production-Grade Capabilities**
- 🔄 **Fault Tolerance** - Circuit breakers, saga compensation
- 📈 **Auto-scaling** - Service discovery, load balancing
- 🛡️ **Security Hardening** - Zero-trust, attribute-based access
- 📊 **Observability** - Distributed tracing, custom metrics

### **Developer Experience** 
- 🎯 **Best Practices** - Proven patterns, enterprise standards
- 🧩 **Modular Design** - Pick and choose components
- 📚 **Comprehensive Examples** - Real-world implementations
- 🔧 **Easy Customization** - Template-based approach

## 💡 **NEXT STEPS**

Bạn có muốn tôi tiếp tục mở rộng thêm các areas khác như:
- **AI/ML Integration Patterns** 
- **Blockchain/Web3 Integration**
- **Real-time Communication** (WebSocket, SSE)
- **Advanced Testing Strategies** (Contract testing, Chaos engineering)
- **DevOps & Infrastructure as Code** patterns

Hay bạn muốn tôi deep-dive vào bất kỳ pattern nào đã trình bày để tạo implementation examples chi tiết hơn? 