package main

import "github.com/gofiber/fiber/v2"

func main() {
	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("hello aaron, jake")
	})

	app.Get("/:name", func(c *fiber.Ctx) error {
        name := c.Params("name")
        return c.SendString("hello: " + name)
    })

	app.Listen(":8000")
}
