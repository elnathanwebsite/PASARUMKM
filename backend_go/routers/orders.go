package routers

import (
	"backend/database"
	"backend/models"

	"github.com/gofiber/fiber/v2"
)

func GetOrderHistory(c *fiber.Ctx) error {
	userID := c.Params("user_id")

	query := "SELECT id, user_id, product_id, store_id, quantity, total_price, status FROM riwayat_pemesanan WHERE user_id = ?"
	rows, err := database.DB.Query(query, userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"detail": err.Error()})
	}
	defer rows.Close()

	var orders []models.OrderModel
	for rows.Next() {
		var o models.OrderModel
		if err := rows.Scan(&o.ID, &o.UserID, &o.ProductID, &o.StoreID, &o.Quantity, &o.TotalPrice, &o.Status); err != nil {
			return c.Status(500).JSON(fiber.Map{"detail": err.Error()})
		}
		orders = append(orders, o)
	}

	if orders == nil {
		orders = []models.OrderModel{}
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   orders,
	})
}

func PlaceOrder(c *fiber.Ctx) error {
	var order models.OrderModel
	if err := c.BodyParser(&order); err != nil {
		return c.Status(400).JSON(fiber.Map{"detail": err.Error()})
	}

	if order.Status == "" {
		order.Status = "pending"
	}

	query := `
		INSERT INTO riwayat_pemesanan (
			id, user_id, product_id, store_id, quantity, total_price, status
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	_, err := database.DB.Exec(query,
		order.ID, order.UserID, order.ProductID, order.StoreID,
		order.Quantity, order.TotalPrice, order.Status,
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"detail": err.Error()})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Order placed successfully",
	})
}

func SetupOrderRoutes(router fiber.Router) {
	router.Get("/history/:user_id", GetOrderHistory)
	router.Post("/checkout", PlaceOrder)
}
