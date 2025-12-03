package controller

import (
	"gobackend/models"
	"strconv"
	"time"

	"gobackend/config"

	"github.com/gofiber/fiber/v2"
	razorpay "github.com/razorpay/razorpay-go"
)

func CreateOrder(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req models.OrderRequest
		if err:=c.BodyParser(&req); err!=nil{
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":"Invalid request body",
			})
		}

		// Initialize razorpay client 
		client := razorpay.NewClient(cfg.RazorpayKeyId, cfg.RazorpayKeySecret)
		data:=map[string]interface{}{
			"amount":req.Amount*100,
			"currency":"INR",//req.Currency,
			"receipt":"order_" + strconv.Itoa(int(time.Now().Unix())),
			
		}

		// create order
		order, err := client.Order.Create(data, nil)
		if err!=nil{
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":"Failed to create order",
			})
		}
		response:=models.OrderResponse{
			Id:order["id"].(string),
			Amount:order["amount"].(float64),
			Currency:order["currency"].(string),
			Status:order["status"].(string),
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
    		"success": true,
    		"data":    response,
})
	}
}