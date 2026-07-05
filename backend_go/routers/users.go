package routers

import (
	"backend/database"
	"backend/models"
	"database/sql"

	"github.com/gofiber/fiber/v2"
)

func GetUser(c *fiber.Ctx) error {
	uid := c.Params("uid")

	query := "SELECT uid, email, displayName, photoURL, phoneNumber FROM users WHERE uid = ?"
	row := database.DB.QueryRow(query, uid)

	var u models.UserModel
	var displayName, photoURL, phoneNumber sql.NullString

	err := row.Scan(&u.UID, &u.Email, &displayName, &photoURL, &phoneNumber)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(fiber.Map{
				"status":  "error",
				"message": "User not found",
			})
		}
		return c.Status(500).JSON(fiber.Map{"detail": err.Error()})
	}

	if displayName.Valid {
		u.DisplayName = displayName.String
	}
	if photoURL.Valid {
		u.PhotoURL = photoURL.String
	}
	if phoneNumber.Valid {
		u.PhoneNumber = phoneNumber.String
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   u,
	})
}

func CreateOrUpdateUser(c *fiber.Ctx) error {
	var user models.UserModel
	if err := c.BodyParser(&user); err != nil {
		return c.Status(400).JSON(fiber.Map{"detail": err.Error()})
	}

	// Check if user exists
	checkQuery := "SELECT uid FROM users WHERE uid = ?"
	var existingUID string
	err := database.DB.QueryRow(checkQuery, user.UID).Scan(&existingUID)

	if err == nil {
		// Update
		query := `
			UPDATE users 
			SET email = ?, displayName = ?, photoURL = ?, phoneNumber = ?
			WHERE uid = ?
		`
		_, err := database.DB.Exec(query, user.Email, user.DisplayName, user.PhotoURL, user.PhoneNumber, user.UID)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"detail": err.Error()})
		}
	} else {
		// Insert
		query := `
			INSERT INTO users (uid, email, displayName, photoURL, phoneNumber) 
			VALUES (?, ?, ?, ?, ?)
		`
		_, err := database.DB.Exec(query, user.UID, user.Email, user.DisplayName, user.PhotoURL, user.PhoneNumber)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"detail": err.Error()})
		}
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "User saved successfully",
	})
}

func SetupUserRoutes(router fiber.Router) {
	router.Get("/:uid", GetUser)
	router.Post("/", CreateOrUpdateUser)
}
