package routers

import (
	"backend/database"
	"backend/models"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"

	"github.com/gofiber/fiber/v2"
)

var (
	LouvinBaseURL = "https://api.louvin.dev"
)

func getLouvinAPIKey() string {
	key := os.Getenv("LOUVIN_API_KEY")
	if key == "" {
		return "lv_e837b09fd975494bb25ab6eaeffef179"
	}
	return key
}

func CreateSubscription(c *fiber.Ctx) error {
	var req models.SubscriptionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"detail": err.Error()})
	}

	payload := map[string]interface{}{
		"amount":         60000,
		"payment_type":   "qris",
		"customer_name":  req.CustomerName,
		"customer_email": req.CustomerEmail,
		"description":    "Upgrade Akun Pro - Pasar UMKM",
		"reference":      "PRO-" + req.UID,
	}

	payloadBytes, _ := json.Marshal(payload)

	httpReq, err := http.NewRequest("POST", LouvinBaseURL+"/create-transaction", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"detail": err.Error()})
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", getLouvinAPIKey())

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"detail": err.Error()})
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return c.Status(500).JSON(fiber.Map{"detail": "Failed to parse Louvin response"})
	}

	if success, ok := data["success"].(bool); !ok || !success {
		errMsg := "Gagal menghubungi Louvin API"
		if msg, ok := data["error"].(string); ok {
			errMsg = msg
		}
		return c.Status(500).JSON(fiber.Map{"detail": errMsg})
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   data,
	})
}

func CheckSubscriptionStatus(c *fiber.Ctx) error {
	transactionID := c.Query("transaction_id")
	uid := c.Query("uid")

	httpReq, err := http.NewRequest("GET", LouvinBaseURL+"/check-status?id="+transactionID, nil)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"detail": err.Error()})
	}
	httpReq.Header.Set("x-api-key", getLouvinAPIKey())

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"detail": err.Error()})
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return c.Status(500).JSON(fiber.Map{"detail": "Failed to parse Louvin response"})
	}

	if success, ok := data["success"].(bool); !ok || !success {
		errMsg := "Gagal cek status transaksi"
		if msg, ok := data["error"].(string); ok {
			errMsg = msg
		}
		return c.Status(500).JSON(fiber.Map{"detail": errMsg})
	}

	txData, _ := data["transaction"].(map[string]interface{})
	status, _ := txData["status"].(string)

	if status == "settled" {
		query := `
			UPDATE users 
			SET isSubscribed = 1, aiLimit = aiLimit + 10, aiSearchLimit = aiSearchLimit + 10, aiDescLimit = aiDescLimit + 10, aiAnalysisLimit = aiAnalysisLimit + 10
			WHERE uid = ?
		`
		_, err := database.DB.Exec(query, uid)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"detail": err.Error()})
		}
		return c.JSON(fiber.Map{
			"status":         "success",
			"payment_status": "settled",
			"message":        "Akun berhasil diupgrade ke PRO",
		})
	}

	return c.JSON(fiber.Map{
		"status":         "success",
		"payment_status": status,
	})
}

func SetupSubscriptionRoutes(router fiber.Router) {
	router.Post("/create", CreateSubscription)
	router.Get("/check-status", CheckSubscriptionStatus)
}
