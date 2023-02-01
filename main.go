package main

import (
	"catalog/api"
	"catalog/config"
	"catalog/repository"
	"catalog/service"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/halilylm/prometheusmiddleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"net/http"
)

func main() {
	registry := prometheus.NewRegistry()
	registry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	registry.MustRegister(collectors.NewGoCollector())

	conf, _ := config.NewConfig("./config/config.yaml")
	repo, _ := repository.NewMongoRepository(conf.Database.URL, conf.Database.DB, conf.Database.Timeout)
	productService := service.NewProductService(repo)
	handler := api.NewHandler(productService)

	app := fiber.New()

	middlewarePath := "/metrics"
	middleware := prometheusmiddleware.NewPrometheusMiddleware(registry, middlewarePath)
	middleware.Use(app)

	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${ip}  ${status} - ${latency} ${method} ${path}\n",
	}))

	app.Get("/products/{code}", handler.Get)
	app.Post("/products", handler.Post)
	app.Delete("/products/{code}", handler.Delete)
	app.Get("/products", handler.GetAll)
	app.Put("/products", handler.Put)
	app.Get("/error", func(ctx *fiber.Ctx) error {
		return ctx.SendStatus(http.StatusInternalServerError)
	})
	app.Listen(":8082")

}
