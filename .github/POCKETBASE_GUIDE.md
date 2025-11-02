# PocketBase Go Quick Reference Guide

**PocketBase v0.31.0** - Embedded database with realtime subscriptions, auth, and admin dashboard.

---

## 📦 **Core Concepts**

### **Initialize App**

```go
import "github.com/pocketbase/pocketbase"

app := pocketbase.New()
// or with config:
app := pocketbase.NewWithConfig(&pocketbase.Config{
    DefaultDataDir: "./pb_data",
})

if err := app.Start(); err != nil {
    log.Fatal(err)
}
```

### **Data Structure**

- **Collections**: Database tables (base, auth, view types)
- **Records**: Rows in collections
- **Fields**: Column definitions (text, number, relation, file, etc.)
- **Migrations**: Versioned schema changes

---

## 🗂️ **Collections**

### **Fetch Collections**

```go
// Find by name or ID
collection, err := app.FindCollectionByNameOrId("articles")

// Find all
allCollections, err := app.FindAllCollections()

// Find by type
authCollections, err := app.FindAllCollections(core.CollectionTypeAuth)
```

### **Create Collection**

```go
collection := core.NewBaseCollection("posts")
// or: core.NewAuthCollection("users")
// or: core.NewViewCollection("stats_view")

// Set CRUD rules
collection.ListRule = types.Pointer("@request.auth.id != ''")
collection.CreateRule = types.Pointer("@request.auth.id = @request.body.author")

// Add fields
collection.Fields.Add(
    &core.TextField{
        Name:     "title",
        Required: true,
        Max:      200,
    },
    &core.RelationField{
        Name:         "author",
        CollectionId: usersCollection.Id,
        Required:     true,
    },
)

// Add index
collection.AddIndex("idx_title", false, "title", "")

// Save
err = app.Save(collection)
```

### **Available Field Types**

- `core.TextField`, `core.EmailField`, `core.URLField`
- `core.NumberField`, `core.BoolField`
- `core.DateField`, `core.AutodateField`
- `core.SelectField`, `core.FileField`
- `core.RelationField`, `core.JSONField`
- `core.EditorField`, `core.GeoPointField`

---

## 📝 **Records**

### **Fetch Records**

```go
// By ID
record, err := app.FindRecordById("articles", "RECORD_ID")

// By field value
record, err := app.FindFirstRecordByData("articles", "slug", "hello-world")

// By filter expression (use {:param} for safety)
record, err := app.FindFirstRecordByFilter(
    "articles",
    "status = 'public' && category = {:cat}",
    dbx.Params{"cat": "tech"},
)

// Multiple records
records, err := app.FindRecordsByFilter(
    "articles",
    "status = 'public'",
    "-published", // sort
    10,           // limit
    0,            // offset
    dbx.Params{},
)

// Auth records
user, err := app.FindAuthRecordByEmail("users", "test@example.com")
user, err := app.FindAuthRecordByToken(token, core.TokenTypeAuth)
```

### **Create Record**

```go
collection, _ := app.FindCollectionByNameOrId("articles")
record := core.NewRecord(collection)

record.Set("title", "Hello World")
record.Set("status", "published")
record.Set("tags", []string{"go", "backend"})

// Field modifiers
record.Set("slug:autogenerate", "post-")

// Save with validation
err = app.Save(record)

// Save without validation
err = app.SaveNoValidate(record)
```

### **Update Record**

```go
record, _ := app.FindRecordById("articles", "RECORD_ID")

record.Set("title", "Updated Title")
record.Set("views+", 1) // increment

err = app.Save(record)
```

### **Delete Record**

```go
record, _ := app.FindRecordById("articles", "RECORD_ID")
err = app.Delete(record)
```

### **Get Field Values**

```go
record.Get("title")              // -> any
record.GetString("title")        // -> string
record.GetInt("views")           // -> int
record.GetBool("published")      // -> bool
record.GetFloat("rating")        // -> float64
record.GetDateTime("created")    // -> types.DateTime
record.GetStringSlice("tags")    // -> []string
```

### **Expand Relations**

```go
// Expand programmatically
errs := app.ExpandRecord(record, []string{"author", "categories"}, nil)

// Access expanded
author := record.ExpandedOne("author")       // -> *core.Record
categories := record.ExpandedAll("categories") // -> []*core.Record
```

---

## 🔄 **Database Operations**

### **Raw SQL Queries**

```go
// Execute (no data returned)
_, err := app.DB().NewQuery("UPDATE articles SET status = 'archived'").Execute()

// One result
type User struct {
    Id   string `db:"id"`
    Name string `db:"name"`
}
user := User{}
err := app.DB().NewQuery("SELECT id, name FROM users WHERE id = 1").One(&user)

// Multiple results
users := []User{}
err := app.DB().NewQuery("SELECT * FROM users LIMIT 10").All(&users)

// With parameters (safe from SQL injection)
err := app.DB().
    NewQuery("SELECT * FROM posts WHERE created >= {:from}").
    Bind(dbx.Params{"from": "2023-01-01"}).
    All(&posts)
```

### **Query Builder**

```go
users := []User{}

err := app.DB().
    Select("id", "email", "created").
    From("users").
    Where(dbx.HashExp{"status": "active"}).
    AndWhere(dbx.Like("email", "example.com")).
    OrderBy("created DESC").
    Limit(50).
    All(&users)
```

### **Common Expressions**

```go
// Hash expression
dbx.HashExp{"status": "active", "verified": true}

// Custom expression
dbx.NewExp("age > {:min}", dbx.Params{"min": 18})

// IN / NOT IN
dbx.In("status", "active", "pending")
dbx.NotIn("role", "admin", "moderator")

// LIKE / NOT LIKE
dbx.Like("name", "john")           // name LIKE "%john%"
dbx.NotLike("email", "spam.com")

// AND / OR / NOT
dbx.And(exp1, exp2)
dbx.Or(exp1, exp2)
dbx.Not(dbx.HashExp{"deleted": true})

// BETWEEN
dbx.Between("age", 18, 65)

// EXISTS
dbx.Exists(subquery)
```

### **Transactions**

```go
err := app.RunInTransaction(func(txApp core.App) error {
    // Use txApp (not app!) for all operations
    record1, _ := txApp.FindRecordById("articles", "ID1")
    record1.Set("status", "published")
    if err := txApp.Save(record1); err != nil {
        return err // rollback
    }

    _, err := txApp.DB().NewQuery("DELETE FROM drafts").Execute()
    return err // nil = commit, error = rollback
})
```

---

## 🪝 **Event Hooks**

### **Hook Structure**

All hooks use: `func(e *EventType) error { return e.Next() }`

### **App Lifecycle Hooks**

```go
// On server start
app.OnServe().BindFunc(func(se *core.ServeEvent) error {
    // Register routes, setup watchers, etc.
    return se.Next()
})

// On app termination
app.OnTerminate().BindFunc(func(e *core.TerminateEvent) error {
    // Cleanup
    return e.Next()
})
```

### **Record Model Hooks**

```go
// Before create (no request access)
app.OnRecordCreate("articles").BindFunc(func(e *core.RecordEvent) error {
    e.Record.Set("slug:autogenerate", "")
    return e.Next()
})

// After successful create
app.OnRecordAfterCreateSuccess("articles").BindFunc(func(e *core.RecordEvent) error {
    // Send notification, update cache, etc.
    return e.Next()
})

// Before update
app.OnRecordUpdate("users").BindFunc(func(e *core.RecordEvent) error {
    // Compare old vs new: e.Record vs e.Record.Original()
    return e.Next()
})

// Before delete
app.OnRecordDelete().BindFunc(func(e *core.RecordEvent) error {
    // Cleanup related data
    return e.Next()
})
```

### **Record Request Hooks** (with request access)

```go
// Intercept create request
app.OnRecordCreateRequest("articles").BindFunc(func(e *core.RecordRequestEvent) error {
    // Access request: e.Auth, e.RequestInfo(), etc.
    if e.Auth == nil {
        return e.UnauthorizedError("Login required", nil)
    }

    e.Record.Set("author", e.Auth.Id)
    return e.Next()
})

// Intercept update request
app.OnRecordUpdateRequest("articles").BindFunc(func(e *core.RecordRequestEvent) error {
    // Prevent non-owners from editing
    if e.Record.GetString("author") != e.Auth.Id {
        return e.ForbiddenError("Not your article", nil)
    }
    return e.Next()
})
```

### **Other Hooks**

- `OnRecordEnrich` - Customize record serialization
- `OnRecordValidate` - Custom validation logic
- `OnRealtimeConnectRequest` - Intercept realtime subscriptions
- `OnMailerSend` - Intercept emails

---

## 🚏 **Routing & Custom Endpoints**

### **Register Routes**

```go
app.OnServe().BindFunc(func(se *core.ServeEvent) error {
    // Simple GET
    se.Router.GET("/hello/{name}", func(e *core.RequestEvent) error {
        name := e.Request.PathValue("name")
        return e.String(200, "Hello "+name)
    })

    // POST with auth
    se.Router.POST("/api/custom", func(e *core.RequestEvent) error {
        if e.Auth == nil {
            return e.UnauthorizedError("Login required", nil)
        }
        return e.JSON(200, map[string]any{"ok": true})
    }).Bind(apis.RequireAuth())

    // Route groups
    g := se.Router.Group("/api/posts")
    g.Bind(apis.RequireAuth()) // group middleware
    g.GET("", listPosts)
    g.GET("/{id}", getPost)
    g.POST("", createPost)
    g.PATCH("/{id}", updatePost)
    g.DELETE("/{id}", deletePost)

    return se.Next()
})
```

### **Request Handlers**

```go
func customHandler(e *core.RequestEvent) error {
    // Path params
    id := e.Request.PathValue("id")

    // Query params
    search := e.Request.URL.Query().Get("search")

    // Headers
    token := e.Request.Header.Get("Authorization")

    // Body (struct or map)
    data := struct {
        Title string `json:"title"`
        Text  string `json:"text"`
    }{}
    if err := e.BindBody(&data); err != nil {
        return e.BadRequestError("Invalid data", err)
    }

    // Auth state
    if e.Auth == nil {
        return e.UnauthorizedError("", nil)
    }
    userID := e.Auth.Id
    isSuperuser := e.HasSuperuserAuth()

    // Responses
    return e.JSON(200, data)
    // or: e.String(200, "text")
    // or: e.HTML(200, "<html>...")
    // or: e.NoContent(204)
    // or: e.Redirect(301, "/other")
}
```

### **Middlewares**

```go
// Global middleware
se.Router.BindFunc(func(e *core.RequestEvent) error {
    // Before request
    start := time.Now()

    err := e.Next() // continue chain

    // After request
    log.Printf("Request took %v", time.Since(start))
    return err
})

// Built-in middlewares
apis.RequireAuth()              // any auth record
apis.RequireAuth("users")       // specific collection
apis.RequireSuperuserAuth()     // superuser only
apis.RequireGuestOnly()         // unauthenticated only
apis.BodyLimit(10 << 20)        // 10MB limit
apis.Gzip()                     // compress response
```

---

## 🔧 **Migrations**

### **Setup**

```go
import "github.com/pocketbase/pocketbase/plugins/migratecmd"

migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
    Automigrate: isGoRun, // auto-create on Dashboard changes (dev only)
})
```

### **Create Migration**

```bash
go run . migrate create "add_posts_collection"
```

### **Migration Template**

```go
// migrations/1234567890_add_posts_collection.go
package migrations

import (
    "github.com/pocketbase/pocketbase/core"
    m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
    m.Register(
        // UP
        func(app core.App) error {
            collection := core.NewBaseCollection("posts")
            collection.Fields.Add(
                &core.TextField{Name: "title", Required: true, Max: 200},
                &core.TextField{Name: "content"},
            )
            return app.Save(collection)
        },
        // DOWN (optional)
        func(app core.App) error {
            collection, _ := app.FindCollectionByNameOrId("posts")
            return app.Delete(collection)
        },
    )
}
```

### **Load Migrations**

```go
// main.go
import _ "yourpackage/migrations"
```

### **Run Migrations**

```bash
go run . serve          # auto-apply on start
go run . migrate up     # manual apply
go run . migrate down 1 # revert last migration
```

---

## 🧪 **Testing**

### **Test Setup**

```go
import "github.com/pocketbase/pocketbase/tests"

func TestCustomEndpoint(t *testing.T) {
    testApp, err := tests.NewTestApp("./test_pb_data")
    if err != nil {
        t.Fatal(err)
    }
    defer testApp.Cleanup()

    // Register your hooks
    bindAppHooks(testApp)

    // Generate auth token
    user, _ := testApp.FindAuthRecordByEmail("users", "test@example.com")
    token, _ := user.NewAuthToken()

    // Test scenarios
    scenarios := []tests.ApiScenario{
        {
            Name:            "guest access denied",
            Method:          http.MethodGet,
            URL:             "/api/protected",
            ExpectedStatus:  401,
            TestAppFactory:  func(t testing.TB) *tests.TestApp { return testApp },
        },
        {
            Name:            "auth user success",
            Method:          http.MethodGet,
            URL:             "/api/protected",
            Headers:         map[string]string{"Authorization": token},
            ExpectedStatus:  200,
            ExpectedContent: []string{"success"},
            TestAppFactory:  func(t testing.TB) *tests.TestApp { return testApp },
        },
    }

    for _, scenario := range scenarios {
        scenario.Test(t)
    }
}
```

---

## 🔑 **Auth & Tokens**

### **Generate Tokens**

```go
// Auth token
token, err := record.NewAuthToken()

// Other token types
verifyToken, _ := record.NewVerificationToken()
resetToken, _ := record.NewPasswordResetToken()
emailChangeToken, _ := record.NewEmailChangeToken("new@example.com")
fileToken, _ := record.NewFileToken()
```

### **Validate Tokens**

```go
user, err := app.FindAuthRecordByToken(token, core.TokenTypeAuth)
```

### **Auth Response Helper**

```go
se.Router.POST("/phone-login", func(e *core.RequestEvent) error {
    // ... validate phone/password ...
    return apis.RecordAuthResponse(e, record, "phone", nil)
})
```

---

## ⚡ **Best Practices**

1. **Always use parameter binding** for user input: `{:param}` not string concatenation
2. **Use transactions** for multi-record operations: `app.RunInTransaction()`
3. **Inside transactions**, always use `txApp` not the original `app`
4. **Use hooks wisely**: Model hooks for business logic, Request hooks for request validation
5. **Validation**: Use `app.Save()` (validates) not `app.SaveNoValidate()` unless needed
6. **Avoid deadlocks**: Don't use global mutex locks inside hooks (can be recursive)
7. **Testing**: Use separate `test_pb_data` directory with pre-seeded data

---

## 📚 **Common Patterns**

### **Custom Record Query**

```go
func FindActiveArticles(app core.App) ([]*core.Record, error) {
    records := []*core.Record{}
    err := app.RecordQuery("articles").
        AndWhere(dbx.HashExp{"status": "active"}).
        OrderBy("published DESC").
        Limit(10).
        All(&records)
    return records, err
}
```

### **Check Record Access**

```go
info, _ := e.RequestInfo()
canAccess, err := app.CanAccessRecord(record, info, record.Collection().ViewRule)
if !canAccess {
    return e.ForbiddenError("", err)
}
```

### **Serve Static Files**

```go
se.Router.GET("/{path...}", apis.Static(os.DirFS("./public"), false))
```

### **Custom Validation**

```go
app.OnRecordValidate("articles").BindFunc(func(e *core.RecordEvent) error {
    title := e.Record.GetString("title")
    if len(title) < 5 {
        return errors.New("title too short")
    }
    return e.Next()
})
```

---

## 🔗 **Resources**

- **Docs**: https://pocketbase.io/docs
- **Go Pkg**: https://pkg.go.dev/github.com/pocketbase/pocketbase
- **GitHub**: https://github.com/pocketbase/pocketbase
- **Admin Dashboard**: http://localhost:8090/\_/

---

**Key Takeaway**: PocketBase = Embedded SQLite + REST API + Realtime + Admin UI + Auth in one Go package. Perfect for building portable backends quickly.
