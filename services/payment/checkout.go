package payment

import (
	"gobackend/config"
	"strconv"

	"github.com/gofiber/fiber/v2"
)


func ShowCheckoutPage(cfg *config.Config) fiber.Handler{
	return func(c *fiber.Ctx) error{
		orderId:=c.Query("orderId")
		amountStr:=c.Query("amount")

		amount , err := strconv.Atoi(amountStr)
		if err!=nil{
			return c.Status(fiber.StatusInternalServerError).JSON("failed to convert amount")
		}

		data:=map[string]interface{}{
			"amount":amount,
			"orderId":orderId,
		}
		return c.JSON(data)
	}
}