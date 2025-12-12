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
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
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

func maxTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

func isUpdateSubscription(user_id string, plan string, sub models.Subscription) bool {

	if sub.Status == "active" && sub.EndAt.After(time.Now().UTC()) {
		// update subscription
		filter := bson.M{"user_id": user_id}

		now := time.Now().UTC()

		startTime := maxTime(sub.EndAt.Add(24*time.Hour), now.Add(24*time.Hour))
		endTime := startTime.Add(30 * 24 * time.Hour)

		update := bson.M{
			"$set": bson.M{
				"start_at": startTime,
				"end_at":   endTime,
				"status":   "active",
			},
		}
		_, err := connect.SubscriptionCollection.UpdateOne(context.TODO(), filter, update)
		if err != nil {
			return false
		}
	}
	return true

}

var Sub models.Subscription

func anyActiveSubscriptionProOrCreator(user_id string) bool {
	filter := bson.D{{Key: "user_id", Value: user_id}}
	err := connect.SubscriptionCollection.FindOne(context.TODO(), filter).Decode(&Sub)
	if err != nil {
		fmt.Println("Failed to fetch subscription, user may not have any subscription")
		return false
	}
	return true
}

func CreatePaymentLink() fiber.Handler {
	return func(c *fiber.Ctx) error {
		fmt.Println("CreatePaymentLink")
		user_id, err := FetchUserId(c)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to fetch user id",
			})
		}
		envs := env.NewEnv()
		var body struct {
			Plan string `json:"plan"`
		}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		if anyActiveSubscriptionProOrCreator(user_id) {
			if Sub.Plan == body.Plan && Sub.Status == "active" && Sub.EndAt.After(time.Now().UTC()) {
				fmt.Println("User already has an active subscription")
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "User already has an active subscription",
				})
			} else {
				isPlanChanged := Sub.Plan != body.Plan
				if isPlanChanged {
					// update subscription
					if !isUpdateSubscription(user_id, body.Plan, Sub) {
						return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
							"error": "Failed to update subscription",
						})
					} else {
						fmt.Println("Subscription updated successfully")
					}
				}
			}
		} else {
			subscription := models.Subscription{
				UserId:      user_id,
				Plan:        body.Plan,
				Status:      "pending",
				StartAt:     time.Now().UTC(),
				EndAt:       time.Now().UTC().Add(time.Hour * 24 * 30),
				AutoRenew:   false,
				TrialEndsAt: time.Now().UTC().Add(time.Hour * 24 * 30),
				UpdatedAt:   time.Now().UTC(),
			}

			_, err = connect.SubscriptionCollection.InsertOne(context.TODO(), subscription)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Failed to create subscription",
				})
			}
		}

		priceMap := map[string]int64{
			"starter": 0,
			"creator": 99 * 100,
			"pro":     499 * 100,
		}

		amount := priceMap[strings.ToLower(body.Plan)]

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
			"callback_url":    "http://localhost:3000/dashboard/payment/subscription/success", //change the domain later
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

// if payment is verified , then update subscription status.
type SubscriptionSucessResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func SubscriptionSuccess() fiber.Handler {
	return func(c *fiber.Ctx) error {
		user_id, err := FetchUserId(c)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to fetch user id",
			})
		}
		envs := env.NewEnv()
		queries := c.Queries()
		params := map[string]interface{}{
			"payment_link_id":           queries["razorpay_payment_link_id"],
			"razorpay_payment_id":       queries["razorpay_payment_id"],
			"payment_link_reference_id": queries["razorpay_payment_link_reference_id"],
			"payment_link_status":       queries["razorpay_payment_link_status"],
		}
		signature := queries["razorpay_signature"]
		secret := envs.RAZORPAY_KEY_SECRET

		// update subscription status pending to active
		if utils.VerifyPaymentLinkSignature(params, signature, secret) {
			_, err = connect.SubscriptionCollection.UpdateOne(context.TODO(), bson.D{{Key: "user_id", Value: user_id}}, bson.D{{Key: "$set", Value: bson.D{{Key: "status", Value: "active"}}}})
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(SubscriptionSucessResponse{
					Success: false,
					Message: "Failed to update subscription status",
				})
			}
			return c.Status(fiber.StatusOK).JSON(SubscriptionSucessResponse{
				Success: true,
				Message: "Subscription updated successfully",
			})
		}

		return c.Status(fiber.StatusUnauthorized).JSON(SubscriptionSucessResponse{
			Success: false,
			Message: "Invalid signature",
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
			UserId:      user_id,
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
			{Key: "user_id", Value: user_id},
			{Key: "status", Value: "pending"},
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
			{Key: "$set", Value: bson.D{
				{Key: "status", Value: "active"},
				{Key: "end_at", Value: time.Now().UTC().Add(duration)},
				{Key: "updated_at", Value: time.Now().UTC()},
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

// Mark Subsc
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

type GetActiveSubscriptionResponse struct {
	Plan   string `json:"plan"`
	Status bool   `json:"status"`
}

func GetActiveSubscription() fiber.Handler {
	return func(c *fiber.Ctx) error {
		user_id, err := FetchUserId(c)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to fetch user id",
			})
		}

		filter := bson.D{
			{Key: "user_id", Value: user_id},
		}

		fmt.Println("GetActiveSubscription: Searching for user_id:", user_id)
		var subscription models.Subscription
		err = connect.SubscriptionCollection.FindOne(context.TODO(), filter).Decode(&subscription)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				fmt.Println("GetActiveSubscription: No subscription found for user_id:", user_id)
				return c.Status(fiber.StatusOK).JSON(GetActiveSubscriptionResponse{
					Plan:   "free", // Or handle as no plan
					Status: false,
				})
			}
			fmt.Println("GetActiveSubscription: Error fetching subscription:", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to fetch subscription details: %v", err),
			})
		}
		if subscription.Status == "active" {
			return c.Status(fiber.StatusOK).JSON(GetActiveSubscriptionResponse{
				Plan:   subscription.Plan,
				Status: true,
			})
		}
		return c.Status(fiber.StatusOK).JSON(GetActiveSubscriptionResponse{
			Plan:   subscription.Plan,
			Status: false,
		})
	}
}
