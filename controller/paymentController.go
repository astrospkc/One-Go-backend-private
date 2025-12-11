package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"gobackend/connect"
	"gobackend/env"
	"gobackend/models"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gobackend/config"

	"github.com/gofiber/fiber/v2"
	razorpay "github.com/razorpay/razorpay-go"
	"github.com/razorpay/razorpay-go/utils"
	"go.mongodb.org/mongo-driver/bson"
)

func CreateOrder(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req models.OrderRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		// Initialize razorpay client
		client := razorpay.NewClient(cfg.RazorpayKeyId, cfg.RazorpayKeySecret)
		data := map[string]interface{}{
			"amount":   req.Amount * 100,
			"currency": "INR", //req.Currency,
			"receipt":  "order_" + strconv.Itoa(int(time.Now().Unix())),
		}

		// create order
		order, err := client.Order.Create(data, nil)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to create order",
			})
		}
		response := models.OrderResponse{
			Id:       order["id"].(string),
			Amount:   order["amount"].(float64),
			Currency: order["currency"].(string),
			Status:   order["status"].(string),
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"success": true,
			"data":    response,
		})
	}
}

func CreatePaymentLink() fiber.Handler {
	return func(c *fiber.Ctx) error {
		envs := env.NewEnv()
		var body struct {
			Plan string `json:"plan"`
		}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}
		priceMap := map[string]int64{
			"starter": 0,
			"creator": 99 * 100,
			"pro":     499 * 100,
		}
		fmt.Println("body.Plan: ", body.Plan)
		amount := priceMap[strings.ToLower(body.Plan)]
		fmt.Println("amount: ", amount)
		payload := map[string]interface{}{
			"amount":      amount,
			"currency":    "INR",
			"description": fmt.Sprintf("Subscription - %s plan", body.Plan),
			"customer": map[string]string{
				"name":  "Test User",
				"email": "user@example.com",
			},
			"notify": map[string]bool{
				"sms":   true,
				"email": true,
			},
			"callback_url":    "http://localhost:3000/payment/subscription/success", //change the domain later
			"callback_method": "get",
		}
		payloadBytes, _ := json.Marshal(payload)

		req, _ := http.NewRequest("POST", "https://api.razorpay.com/v1/payment_links", bytes.NewBuffer(payloadBytes))
		req.Header.Add("Content-Type", "application/json")
		req.SetBasicAuth(envs.RAZORPAY_KEY_ID, envs.RAZORPAY_KEY_SECRET)
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
			"data":    razorResponse,
		})

	}
}

// razorpay_payment_id=pay_RpA339YQ8jrMIk&
// razorpay_payment_link_id=plink_RpA2eytvDXpT4j&
// razorpay_payment_link_reference_id=&razorpay_payment_link_status=paid&
// razorpay_signature=9323c77f4e18b8fd41d7c25fa37db7c10bdf80120429ec979cbed841d602d918

func SubscriptionSuccess() fiber.Handler {
	return func(c *fiber.Ctx) error {
		envs := env.NewEnv()
		queries := c.Queries()

		fmt.Println("queries: ", queries)
		params := map[string]interface{}{
			"payment_link_id":           queries["razorpay_payment_link_id"],
			"razorpay_payment_id":       queries["razorpay_payment_id"],
			"payment_link_reference_id": queries["razorpay_payment_link_reference_id"],
			"payment_link_status":       "paid",
		}
		signature := queries["razorpay_signature"]
		secret := envs.RAZORPAY_KEY_SECRET

		if utils.VerifyPaymentLinkSignature(params, signature, secret) {
			return c.JSON(fiber.Map{
				"message": "Payment verified",
				"success": true,
			})
		}

		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Invalid signature",
			"success": false,
		})

	}
}

// insert subscription pending . if free its active
func CreatePendingSubscription() fiber.Handler {
	return func(c *fiber.Ctx) error {
		now := time.Now()
		var body struct {
			Plan string `bson:"plan" json:"plan"`
		}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}
		var status string
		var endtime time.Time
		if body.Plan == "free" {
			status = "active"
			endtime = now.Add(time.Hour * 24 * 90)
		} else {
			status = "pending"
			endtime = now.Add(time.Hour * 24 * 30)
		}

		user_id, err := FetchUserId(c)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to fetch user id",
			})
		}
		subscription := models.Subscription{
			UserID:      user_id,
			Plan:        body.Plan,
			Status:      status,
			StartAt:     now,
			EndAt:       endtime,
			AutoRenew:   false,
			TrialEndsAt: endtime,
			UpdatedAt:   time.Now().UTC(),
		}
		result, err := connect.SubscriptionCollection.InsertOne(context.TODO(), subscription)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Failed to insert subscri[tion details",
			})
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"success": true,
			"data":    result,
		})
	}
}

func ActivateSubscription() fiber.Handler {
	return func(c *fiber.Ctx) error {
		user_id, err := FetchUserId(c)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to fetch user id",
			})
		}
		var sub models.Subscription
		filter := bson.D{
			{"user_id", user_id},
			{"status", "pending"},
		}
		result := connect.SubscriptionCollection.FindOne(context.TODO(), filter)
		if err := result.Decode(&sub); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to fetch subscription details",
			})
		}
		duration := time.Hour * 24 * 30
		if sub.Plan == "free" {
			duration = time.Hour * 24 * 90
		}
		// find the subscription and see the plan,
		// if any plan just active update the status to active

		updatePayload := bson.D{
			{"$set", bson.D{
				{"status", "active"},
				{"end_at", time.Now().UTC().Add(duration)},
				{"updated_at", time.Now().UTC()},
			}},
		}
		_, err = connect.SubscriptionCollection.UpdateOne(context.TODO(), filter, updatePayload)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to update subscription status",
			})
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"success": true,
		})
	}
}

func MarkSubscriptionFailed() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"success": true,
		})
	}
}

func UpdateAutoRenew() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"success": true,
		})
	}
}

func GetActiveSubscription() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"success": true,
		})
	}
}
