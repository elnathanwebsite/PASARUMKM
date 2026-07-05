package routers

import (
	"backend/database"
	"backend/models"
	"database/sql"

	"github.com/gofiber/fiber/v2"
)

func GetProducts(c *fiber.Ctx) error {
	uid := c.Query("uid")
	var query string
	var rows *sql.Rows
	var err error

	if uid != "" {
		query = "SELECT * FROM products WHERE uid = ?"
		rows, err = database.DB.Query(query, uid)
	} else {
		query = "SELECT * FROM products"
		rows, err = database.DB.Query(query)
	}

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"detail": err.Error()})
	}
	defer rows.Close()

	var products []models.ProductModel
	for rows.Next() {
		var p models.ProductModel
		if err := rows.Scan(
			&p.ID, &p.UID, &p.Name, &p.Price, &p.Description, &p.PhotoURL,
			&p.Stock, &p.Category, &p.IsQrisActive, &p.IsPaypalActive,
			&p.ShippingJNE, &p.ShippingJNT, &p.ShippingSicepat, &p.ShippingPos,
			&p.ShippingInstant, &p.ShippingPickup,
		); err != nil {
			// In case the DB has fewer/more columns, we might need a more flexible scanner 
			// but assuming schema matches exactly what was in Python.
			// Let's use column pointers map for safety in SQLite
			return c.Status(500).JSON(fiber.Map{"detail": "Error scanning row: " + err.Error()})
		}
		products = append(products, p)
	}

	if products == nil {
		products = []models.ProductModel{}
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   products,
	})
}

func CreateProduct(c *fiber.Ctx) error {
	var product models.ProductModel
	if err := c.BodyParser(&product); err != nil {
		return c.Status(400).JSON(fiber.Map{"detail": err.Error()})
	}

	query := `
		INSERT INTO products (
			id, uid, name, price, description, photoUrl, stock, category,
			isQrisActive, isPaypalActive, shippingJNE, shippingJNT,
			shippingSicepat, shippingPos, shippingInstant, shippingPickup
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := database.DB.Exec(query,
		product.ID, product.UID, product.Name, product.Price, product.Description,
		product.PhotoURL, product.Stock, product.Category, product.IsQrisActive,
		product.IsPaypalActive, product.ShippingJNE, product.ShippingJNT,
		product.ShippingSicepat, product.ShippingPos, product.ShippingInstant, product.ShippingPickup,
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"detail": err.Error()})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Product created successfully",
	})
}

func DeleteProduct(c *fiber.Ctx) error {
	productID := c.Params("product_id")
	query := "DELETE FROM products WHERE id = ?"
	_, err := database.DB.Exec(query, productID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"detail": err.Error()})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Product deleted successfully",
	})
}

func SetupProductRoutes(router fiber.Router) {
	router.Get("/", GetProducts)
	router.Post("/", CreateProduct)
	router.Delete("/:product_id", DeleteProduct)
}
