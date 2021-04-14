package router

import (
	"github.com/NikSchaefer/go-fiber/handlers"
	"github.com/NikSchaefer/go-fiber/middleware"
	"github.com/gofiber/fiber/v2"
)

func Initalize(router *fiber.App) {

	router.Use(middleware.Security)

	router.Get("/", func(c *fiber.Ctx) error {
		return c.Status(200).SendString("Hello, World!")
	})

	router.Use(middleware.Json)

	users := router.Group("/users")
	users.Post("/", handlers.CreateUser)
	users.Delete("/", middleware.Authenticated, handlers.DeleteUser)
	users.Put("/", middleware.Authenticated, handlers.ChangePassword)
	users.Get("/", middleware.Authenticated, handlers.GetUserInfo)
	users.Post("/login", handlers.Login)
	users.Delete("/logout", handlers.Logout)

	products := router.Group("/products", middleware.Authenticated)
	products.Post("/", handlers.CreateProduct)
	products.Get("/", handlers.GetProducts)
	products.Delete("/:id", handlers.DeleteProduct)
	products.Get("/:id", handlers.GetProductById)
	products.Put("/:id", handlers.UpdateProduct)

	router.Use(func(c *fiber.Ctx) error {
		return c.SendStatus(404) // => 404 "Not Found"
	})

}
