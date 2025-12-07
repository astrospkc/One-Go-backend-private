package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gobackend/env"
	"gobackend/models"
	"net/http"
	"strconv"
	"strings"
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

func CreatePaymentLink() fiber.Handler{
	return func (c *fiber.Ctx) error{
		envs:=env.NewEnv()
		var body struct{
			Plan string `json:"plan"`
		}
		if err:= c.BodyParser(&body); err!=nil{
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":"Invalid request body",
			})
		}
		priceMap:=map[string]int64{
			"starter":0,
			"creator":99*100,
			"pro":499*100,
			
		}
		fmt.Println("body.Plan: ", body.Plan)
		amount:=priceMap[strings.ToLower(body.Plan)]
		fmt.Println("amount: ", amount)
		payload := map[string]interface{}{
        "amount":   amount,
        "currency": "INR",
        "description": fmt.Sprintf("Subscription - %s plan", body.Plan),
        "customer": map[string]string{
            "name": "Test User",
            "email": "user@example.com",
        },
        "notify": map[string]bool{
            "sms": true,
            "email": true,
        },
        "callback_url": "http://localhost:3000/subscription/success",//change the domain later
        "callback_method": "get",
    }
	payloadBytes,_:=json.Marshal(payload)

	req,_:=http.NewRequest("POST", "https://api.razorpay.com/v1/payment_links",bytes.NewBuffer(payloadBytes))
	req.Header.Add("Content-Type","application/json")
	req.SetBasicAuth(envs.RAZORPAY_KEY_ID,envs.RAZORPAY_KEY_SECRET)
 	client := &http.Client{}
    resp, err := client.Do(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Payment link creation failed",
		})
	}
	defer resp.Body.Close()
		
	var razorResponse map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&razorResponse)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data": razorResponse,
	})


	}
}


func 