# bettermux | 基于 Go 标准库 http.ServeMux 的分组路由与中间件框架

*纯封装 Go 标准库 net/http 的 ServeMux, 仅此而已*

## 开始使用

``` sh
go get github.com/UnknownMp391/bettermux
```

要求 Go 1.22+ (使用增强路由特性, 如方法路由与路径参数)

## 特性

### 方法路由 Get / Post / Put / Delete / Patch / Options

基于 Go 1.22+ 的增强路由, 内部以 `"METHOD /path"` 形式注册 pattern, 支持 `{name}` 路径参数(通过 `r.PathValue` 读取)

``` go
mux := bettermux.NewBetterMux()

mux.Get("/users/{id}", func(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "GET user %s", r.PathValue("id"))
})

mux.Post("/users", func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusCreated)
})

mux.Put("/users/{id}", func(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "PUT user %s", r.PathValue("id"))
})

mux.Patch("/users/{id}", func(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "PATCH user %s", r.PathValue("id"))
})

mux.Delete("/users/{id}", func(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "DELETE user %s", r.PathValue("id"))
})

mux.Options("/users/{id}", func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusNoContent)
})

http.ListenAndServe(":8080", mux)
```

### 分组路由 Route 与 Mount

`Route` 创建独立的子 mux 并挂载到指定前缀下, 子路由拥有自己的命名空间, 可嵌套使用

`Mount` 将任意 `http.Handler` 挂载到前缀下, 并自动处理 `StripPrefix`

``` go
mux := bettermux.NewBetterMux()

mux.Route("/api/", func(api *bettermux.BetterMux) {
    api.Get("/users", func(w http.ResponseWriter, r *http.Request) {
        // GET /api/users
    })
    api.Get("/users/{id}", func(w http.ResponseWriter, r *http.Request) {
        // GET /api/users/42
    })
})

// 静态文件: /static/css/app.css → ./static/css/app.css
mux.Mount("/static", http.FileServer(http.Dir("./static")))

http.ListenAndServe(":8080", mux)
```

### 中间件 With 与 Handle

`With` 声明中间件, 可链式调用, 先声明的先执行 (最外层)

通过 `Handle` / `HandleFunc` 或任意方法路由注册时, 自动将整条中间件链应用到 handler 上

``` go
mux := bettermux.NewBetterMux()

logged := mux.With(func(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        fmt.Println(">>", r.Method, r.URL.Path)
        next.ServeHTTP(w, r)
    })
})

// 链式中间件: 先声明的 Logging 先执行, 再执行 Auth, 最后到达 handler
app := logged.With(func(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.Header.Get("Authorization") == "" {
            http.Error(w, "unauthorized", http.StatusUnauthorized)
            return
        }
        next.ServeHTTP(w, r)
    })
})

app.Get("/hello", func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("Hello with middleware"))
})

http.ListenAndServe(":8080", app)
```