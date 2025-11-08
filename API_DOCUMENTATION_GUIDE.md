# API Documentation Guide - Swagger/OpenAPI Integration

## 🎉 What We Accomplished

We successfully integrated **Swagger/OpenAPI documentation** into your DoIt API using `swaggo/swag`. Your API now has beautiful, interactive documentation that developers can use to explore and test your endpoints!

---

## 📚 What Was Added

### 1. **Dependencies Installed**

```bash
github.com/swaggo/swag/cmd/swag    # CLI tool for generating docs
github.com/swaggo/http-swagger      # Swagger UI HTTP handler
github.com/swaggo/files             # Embedded Swagger UI assets
```

### 2. **New Files Created**

#### `/internal/web/swagger_models.go`

Standardized response models for consistent API documentation:

- `LoginResponse` - Login endpoint response with user and tokens
- `TokenPair` - JWT access and refresh token structure
- `RefreshResponse` - Refresh token endpoint response
- `ErrorResponse` - Standard error response format
- `SuccessResponse` - Generic success response
- `MessageResponse` - Simple message response

#### `/docs` Directory (Auto-generated)

- `docs.go` - Go package with embedded OpenAPI spec
- `swagger.json` - OpenAPI 2.0 spec in JSON format
- `swagger.yaml` - OpenAPI 2.0 spec in YAML format

### 3. **Enhanced Files**

#### `/cmd/doit/main.go`

Added comprehensive API documentation at package level:

- API metadata (title, version, contact)
- Authentication flow description
- Security scheme definition (JWT Bearer tokens)
- Base path and host configuration

#### `/api/v1/auth/auth_dto.go`

Enhanced DTOs with Swagger annotations:

- Field-level descriptions
- Example values for each field
- Enum values for role field

#### `/internal/model/user.go`

Enhanced User model with detailed field documentation:

- UUID format examples
- Timestamp format examples
- Enum values for roles

#### `/api/v1/auth/auth_handler.go`

Documented all authentication endpoints:

- **Register** - User registration with validation
- **Login** - Authentication with rate limiting
- **Refresh** - Token refresh with rotation
- **Logout** - Single device logout
- **LogoutAll** - Multi-device logout

#### `/api/server.go`

Added Swagger UI route:

- Accessible at `/swagger/index.html`
- Configured with deep linking and list expansion

#### `/Makefile`

Added convenient commands:

- `make swagger` - Generate documentation
- `make swagger-fmt` - Format Swagger comments
- `make install-swag` - Install swag CLI

---

## 🚀 How to Use

### **Step 1: Start Your Database**

```bash
# Start PostgreSQL (Docker)
docker start your-postgres-container

# OR start from scratch
make dev-db
```

### **Step 2: Start Your Application**

```bash
make run

# OR
go run ./cmd/doit/main.go
```

### **Step 3: Access Swagger UI**

Open your browser and navigate to:

```
http://localhost:8080/swagger/index.html
```

### **Step 4: Try Out Endpoints**

#### **Without Authentication:**

1. Click on **"POST /auth/register"**
2. Click **"Try it out"**
3. Edit the JSON request body
4. Click **"Execute"**
5. See the response!

#### **With Authentication:**

1. First, register or login to get tokens
2. Click the **"Authorize"** button at the top (🔓 icon)
3. Enter: `Bearer {your-access-token}`
4. Click **"Authorize"**
5. Now try protected endpoints like `/healthcheck`!

---

## 📖 Swagger Annotations Explained

### **Package-Level Documentation** (main.go)

```go
// Package main DoIt API
//
// RESTful API description
//
//	Schemes: http, https
//	Host: localhost:8080
//	BasePath: /
//	Version: 1.0.0
//
//	SecurityDefinitions:
//	bearer:
//	  type: apiKey
//	  name: Authorization
//	  in: header
//
// swagger:meta
package main
```

### **Endpoint Documentation** (handlers)

```go
// Register godoc
// @Summary      Short one-line description
// @Description  Detailed multi-line description
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request body RegisterInput true "Description"
// @Success      200 {object} model.User "Success description"
// @Failure      400 {object} web.ErrorResponse "Error description"
// @Router       /auth/register [post]
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) error {
    // ... handler code
}
```

### **Model Documentation** (structs)

```go
// User represents the domain model
// @Description Detailed user model description
type User struct {
    // Field description
    ID uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`

    // Another field with enum
    Role UserRole `json:"role" example:"user" enums:"user,admin,moderator"`
}
```

---

## 🔄 Workflow: Adding New Endpoints

### **1. Create Your Handler**

```go
// GetTodos godoc
// @Summary      List user's todos
// @Description  Get all todos for the authenticated user
// @Tags         Todos
// @Accept       json
// @Produce      json
// @Security     bearer
// @Param        status query string false "Filter by status" Enums(pending,completed)
// @Success      200 {array} model.Todo "List of todos"
// @Failure      401 {object} web.ErrorResponse "Unauthorized"
// @Router       /v1/todos [get]
func (h *Handler) GetTodos(w http.ResponseWriter, r *http.Request) error {
    // Implementation
}
```

### **2. Create DTOs with Examples**

```go
// CreateTodoInput represents todo creation request
type CreateTodoInput struct {
    // Todo title
    Title string `json:"title" validate:"required" example:"Buy groceries"`

    // Todo description
    Description string `json:"description" example:"Milk, bread, eggs"`
}
```

### **3. Regenerate Documentation**

```bash
make swagger
```

### **4. Test in Swagger UI**

- Refresh browser
- See your new endpoint
- Try it out!

---

## 🎨 Best Practices

### **1. Be Descriptive**

```go
// ❌ Bad
// @Summary Login

// ✅ Good
// @Summary Authenticate user and receive JWT tokens
// @Description Validates credentials and returns access token (15min) and refresh token (7 days)
```

### **2. Document All Response Codes**

```go
// @Success 200 {object} User "Success"
// @Failure 400 {object} ErrorResponse "Bad request"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal error"
```

### **3. Provide Realistic Examples**

```go
// ✅ Good
Email string `json:"email" example:"john.doe@example.com"`

// ❌ Not helpful
Email string `json:"email" example:"email"`
```

### **4. Use Tags to Group Endpoints**

```go
// @Tags Authentication  // Groups all auth endpoints
// @Tags Todos           // Groups all todo endpoints
// @Tags Admin           // Groups all admin endpoints
```

### **5. Document Rate Limits**

```go
// @Description Rate limit: 5 requests per minute per IP
```

### **6. Use Security Annotations**

```go
// @Security bearer  // For protected endpoints
```

---

## 🐛 Troubleshooting

### **Issue: "swag command not found"**

```bash
make install-swag
# OR
go install github.com/swaggo/swag/cmd/swag@latest
```

### **Issue: "docs package not found"**

```bash
make swagger
# This generates the docs package
```

### **Issue: "Example value type error"**

For `map` or `interface{}` types, use:

```go
Metadata map[string]interface{} `json:"metadata" swaggertype:"object"`
```

### **Issue: "Swagger UI not showing"**

1. Check server is running: `curl http://localhost:8080/swagger/doc.json`
2. Verify docs were generated: `ls docs/`
3. Check logs for errors

### **Issue: "Changes not appearing"**

1. Regenerate docs: `make swagger`
2. Restart server
3. Hard refresh browser (Cmd+Shift+R / Ctrl+Shift+R)

---

## 📝 Makefile Commands

```bash
# Generate Swagger documentation
make swagger

# Format Swagger comments
make swagger-fmt

# Install swag CLI
make install-swag

# View all available commands
make help
```

---

## 🌟 Next Steps

### **1. Document Remaining Endpoints**

- Todo endpoints (CRUD operations)
- User endpoints
- Admin endpoints

### **2. Add Query Parameter Examples**

```go
// @Param limit query int false "Limit" default(10) minimum(1) maximum(100)
// @Param offset query int false "Offset" default(0)
// @Param search query string false "Search query"
```

### **3. Add Request Headers**

```go
// @Param X-Request-ID header string false "Request ID for tracking"
```

### **4. Document File Uploads** (if needed)

```go
// @Param file formData file true "File to upload"
```

### **5. Add API Versioning**

When you create v2 endpoints:

```go
// @Router /v2/auth/register [post]
```

### **6. Export OpenAPI Spec**

Your `docs/swagger.json` can be:

- Imported into Postman
- Used with API testing tools
- Shared with frontend developers
- Published on API documentation platforms

---

## 📚 Resources

### **Swag Documentation**

- [Swag GitHub](https://github.com/swaggo/swag)
- [Declarative Comments Format](https://github.com/swaggo/swag#declarative-comments-format)
- [API Operation Annotations](https://github.com/swaggo/swag#api-operation)

### **OpenAPI Specification**

- [OpenAPI 2.0 Spec](https://swagger.io/specification/v2/)
- [Swagger Editor](https://editor.swagger.io/) - Validate your spec

---

## ✅ Summary

You now have:

- ✅ Fully documented authentication endpoints
- ✅ Interactive Swagger UI at `/swagger/index.html`
- ✅ Standardized response models
- ✅ Easy workflow for adding new endpoints
- ✅ Makefile commands for convenience
- ✅ Best practices and examples

**Your API documentation is now production-ready!** 🎉

Developers can now:

- Explore your API without reading code
- Test endpoints directly in the browser
- See request/response examples
- Understand authentication flow
- Copy example requests

---

## 📞 Questions?

If you encounter issues or have questions:

1. Check the troubleshooting section above
2. Review the Swag documentation
3. Examine the examples in your auth handlers
4. Look at the generated `docs/swagger.json`

Happy documenting! 📖✨
