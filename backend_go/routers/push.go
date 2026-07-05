package routers

import (
	"backend/database"
	"backend/models"
	"encoding/json"
	"fmt"

	webpush "github.com/SherClockHolmes/webpush-go"
	"github.com/gofiber/fiber/v2"
)

func getOrCreateVapidKeys() (string, string, error) {
	var pubKey, privKey string

	// Check if keys exist
	queryPub := "SELECT value FROM settings WHERE key = 'vapid_public_key'"
	err := database.DB.QueryRow(queryPub).Scan(&pubKey)
	if err == nil {
		queryPriv := "SELECT value FROM settings WHERE key = 'vapid_private_key'"
		err = database.DB.QueryRow(queryPriv).Scan(&privKey)
		if err == nil {
			return pubKey, privKey, nil
		}
	}

	// Generate new keys
	privateKey, publicKey, err := webpush.GenerateVAPIDKeys()
	if err != nil {
		return "", "", err
	}

	// Save to DB
	insertQuery := "INSERT OR REPLACE INTO settings (key, value) VALUES (?, ?)"
	_, err = database.DB.Exec(insertQuery, "vapid_public_key", publicKey)
	if err != nil {
		return "", "", err
	}
	_, err = database.DB.Exec(insertQuery, "vapid_private_key", privateKey)
	if err != nil {
		return "", "", err
	}

	return publicKey, privateKey, nil
}

func GetVapidPublicKey(c *fiber.Ctx) error {
	pub, _, err := getOrCreateVapidKeys()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"detail": err.Error()})
	}

	return c.JSON(fiber.Map{"public_key": pub})
}

func SendPushNotification(c *fiber.Ctx) error {
	var payload models.PushPayload
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{"detail": err.Error()})
	}
	if payload.URL == "" {
		payload.URL = "/dashboard.html"
	}

	_, privateKey, err := getOrCreateVapidKeys()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"detail": err.Error()})
	}

	query := "SELECT endpoint, p256dh, auth FROM push_subscriptions WHERE user_id = ?"
	rows, err := database.DB.Query(query, payload.UserID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"detail": err.Error()})
	}
	defer rows.Close()

	sentCount := 0
	deletedCount := 0

	notificationData, _ := json.Marshal(map[string]string{
		"title": payload.Title,
		"body":  payload.Body,
		"url":   payload.URL,
	})

	for rows.Next() {
		var endpoint, p256dh, auth string
		if err := rows.Scan(&endpoint, &p256dh, &auth); err != nil {
			continue
		}

		sub := &webpush.Subscription{
			Endpoint: endpoint,
			Keys: webpush.Keys{
				P256dh: p256dh,
				Auth:   auth,
			},
		}

		resp, err := webpush.SendNotification(notificationData, sub, &webpush.Options{
			Subscriber:      "mailto:admin@pasarumkm.com",
			VAPIDPrivateKey: privateKey,
		})

		if err != nil {
			fmt.Printf("Failed to send WebPush to endpoint %s: %v\n", endpoint, err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == 404 || resp.StatusCode == 410 {
			// Delete expired subscription
			delQuery := "DELETE FROM push_subscriptions WHERE endpoint = ?"
			_, delErr := database.DB.Exec(delQuery, endpoint)
			if delErr != nil {
				fmt.Printf("Failed to delete expired subscription: %v\n", delErr)
			} else {
				deletedCount++
			}
		} else {
			sentCount++
		}
	}

	return c.JSON(fiber.Map{
		"status":                "success",
		"sent_count":            sentCount,
		"deleted_expired_count": deletedCount,
	})
}

func SetupPushRoutes(router fiber.Router) {
	router.Get("/vapid-public-key", GetVapidPublicKey)
	router.Post("/send", SendPushNotification)
}
