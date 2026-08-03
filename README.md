# bettermux

一个完全基于 Go 自带的 http.ServeMux 实现分组路由与中间件包装的低开销路由框架

**0 运行时开销, 本模块逻辑只在路由构建时运行**

API 类似 [chi](https://github.com/go-chi/chi)