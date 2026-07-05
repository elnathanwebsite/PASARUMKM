package routers

import (
	"backend/database"
	"time"

	"github.com/gofiber/fiber/v2"
)

type Transaction struct {
	ID            string    `json:"id"`
	BuyerName     string    `json:"buyer_name"`
	SellerName    *string   `json:"seller_name"`
	ProductName   string    `json:"product_name"`
	Qty           int       `json:"qty"`
	TotalPrice    float64   `json:"total_price"`
	Status        string    `json:"status"`
	PaymentMethod string    `json:"payment_method"`
	CreatedAt     time.Time `json:"created_at"`
}

func GetAllTransactions(c *fiber.Ctx) error {
	query := `
		SELECT 
			r.id,
			r.buyer_name,
			u.businessName as seller_name,
			r.product_name,
			r.qty,
			r.total_price,
			r.status,
			r.payment_method,
			r.created_at
		FROM riwayat_pemesanan r
		LEFT JOIN users u ON r.seller_uid = u.uid
		ORDER BY r.created_at DESC
	`

	rows, err := database.DB.Query(query)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"detail": err.Error()})
	}
	defer rows.Close()

	var transactions []Transaction
	for rows.Next() {
		var t Transaction
		var createdAt string
		err := rows.Scan(
			&t.ID,
			&t.BuyerName,
			&t.SellerName,
			&t.ProductName,
			&t.Qty,
			&t.TotalPrice,
			&t.Status,
			&t.PaymentMethod,
			&createdAt,
		)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"detail": err.Error()})
		}
		
		t.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		transactions = append(transactions, t)
	}

	if transactions == nil {
		transactions = []Transaction{}
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   transactions,
	})
}

func SetupAdminRoutes(router fiber.Router) {
	router.Get("/transactions", GetAllTransactions)
}
