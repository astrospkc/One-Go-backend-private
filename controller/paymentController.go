package controller

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
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
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)



type RazorpayEvent struct {
	Event string `json:"event"`
	Payload struct {
		Payment struct {
			Entity struct {
				Id        string `json:"id"`
				Status    string `json:"status"`
				Amount    int64  `json:"amount"`
				Currency  string `json:"currency"`
				Notes map[string]string	`json:"notes"`
			} `json:"entity"`
		} `json:"payment"`

		Refund struct {
			Entity struct {
				Id       string `json:"id"`
				PaymentId string `json:"payment_id"`
				Status   string `json:"status"`
				Amount   int64  `json:"amount"`
				Notes map[string]string	`json:"notes"`
			} `json:"entity"`
		} `json:"refund"`
	} `json:"payload"`

	
}


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

func isUpdateSubscription(user_id string, plan string, sub models.Subscription) bool {

	fmt.Println("subscription prev: ", sub,sub.Id, sub.Status, sub.EndAt)
	if sub.Status == "active" && sub.EndAt.After(time.Now().UTC()) {
		// update the previous subscription with finished status and create new subscription
		filterUpdate := bson.M{
			"user_id": user_id,
			"id":      sub.Id,
		}
		now := time.Now().UTC()

		startTime := now.Add(24 * time.Hour)
		endTime := startTime.Add(30 * 24 * time.Hour)

		update := bson.M{
			"$set": bson.M{
				"status": "finished",
			},
		}
		_, err := connect.SubscriptionCollection.UpdateOne(context.TODO(), filterUpdate, update)
		if err != nil {
			return false
		}
		plan_update := "pending"
		if plan == "starter" {
			plan_update = "active"
		}
		newSubscription := models.Subscription{
			Id:          primitive.NewObjectID().Hex(),
			UserId:      user_id,
			Plan:        plan,
			Status:      plan_update,
			StartAt:     now,
			EndAt:       endTime,
			AutoRenew:   false,
			TrialEndsAt: endTime,
			UpdatedAt:   now,
		}
		_, err = connect.SubscriptionCollection.InsertOne(context.TODO(), newSubscription)
		if err != nil {
			return false
		}
	}
	return true

}


func anyActiveSubscriptionProOrCreator(user_id string) (bool, models.Subscription) {
	var subscription models.Subscription

	filter := bson.D{{Key: "user_id", Value: user_id},{Key:"status",Value:"active"}}
	err := connect.SubscriptionCollection.FindOne(context.TODO(), filter).Decode(&subscription)
	if err != nil {
		fmt.Println("Failed to fetch subscription, user may not have any subscription")
		return false, models.Subscription{}
	}
	// if Sub.Plan == "starter" {
	// 	return false
	// }
	return true, subscription
}

type CreatePaymentLinkResponse struct {
	Success bool                   `json:"success"`
	Data    map[string]interface{} `json:"data"`
	Message string                 `json:"message"`
}

func CreatePaymentLink() fiber.Handler {
	return func(c *fiber.Ctx) error {

		user_id, err := FetchUserId(c)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(CreatePaymentLinkResponse{
				Success: false,
				Data:    nil,
				Message: "Failed to fetch user id",
			})
		}
		envs := env.NewEnv()
		var body struct {
			Plan string `json:"plan"`
		}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(CreatePaymentLinkResponse{
				Success: false,
				Data:    nil,
				Message: "Invalid request body",
			})
		}

		
		isActive,subscription := anyActiveSubscriptionProOrCreator(user_id)

		if isActive {
			if subscription.Plan == body.Plan && subscription.Status == "active" && subscription.EndAt.After(time.Now().UTC()) {
				fmt.Println("User already has an active subscription")
				return c.Status(fiber.StatusBadRequest).JSON(CreatePaymentLinkResponse{
					Success: false,
					Data:    nil,
					Message: "User already has an active subscription",
				})
			} else {
				isPlanChanged := subscription.Plan != body.Plan
				fmt.Println("isPlanChanged: ", isPlanChanged)
			
				if isPlanChanged {
					// update subscription
					if !isUpdateSubscription(user_id, body.Plan, subscription) {
						return c.Status(fiber.StatusInternalServerError).JSON(CreatePaymentLinkResponse{
							Success: false,
							Data:    nil,
							Message: "Failed to update subscription",
						})
					} else {

						// in place of short_url for starter - keep the frontend redirect url
						response := map[string]interface{}{
							"short_url": `http://localhost:3000/dashboard`,
						}
						if body.Plan=="starter"{
							return c.Status(fiber.StatusOK).JSON(CreatePaymentLinkResponse{
								Success: true,
								Data:    response,
								Message: "Subscription updated successfully",
							})
						}
						fmt.Println("Subscription updated successfully")
					}
				}
			}
		} else {
			status:="pending"
			if body.Plan=="starter"{
				status="active"
			}
			subscription := models.Subscription{
				Id:          primitive.NewObjectID().Hex(),
				UserId:      user_id,
				Plan:        body.Plan,
				Status:      status,
				StartAt:     time.Now().UTC(),
				EndAt:       time.Now().UTC().Add(time.Hour * 24 * 30),
				AutoRenew:   false,
				TrialEndsAt: time.Now().UTC().Add(time.Hour * 24 * 30),
				UpdatedAt:   time.Now().UTC(),
			}

			_, err = connect.SubscriptionCollection.InsertOne(context.TODO(), subscription)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(CreatePaymentLinkResponse{
					Success: false,
					Data:    nil,
					Message: "Failed to create subscription",
				})
			}
		}

		priceMap := map[string]int64{
			"starter": 0,
			"creator": 99 * 100,
			"pro":     299 * 100,
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
			"notes":map[string]string{
				"user_id":user_id,
				"subscription_id":subscription.Id,
				"sub_payment_id":"",
			},
		}
		payloadBytes, _ := json.Marshal(payload)

		req, _ := http.NewRequest("POST", "https://api.razorpay.com/v1/payment_links", bytes.NewBuffer(payloadBytes))
		req.Header.Add("Content-Type", "application/json")
		req.SetBasicAuth(envs.RAZORPAY_KEY_ID, envs.RAZORPAY_KEY_SECRET)
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(CreatePaymentLinkResponse{
				Success: false,
				Data:    nil,
				Message: "Payment link creation failed",
			})
		}
		defer resp.Body.Close()

		var razorResponse map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&razorResponse)
		return c.Status(fiber.StatusOK).JSON(CreatePaymentLinkResponse{
			Success: true,
			Data:    razorResponse,
			Message: "Payment link created successfully",
		})

	}
}

// if payment is verified , then update subscription status.
type SubscriptionSucessResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// in place of this webhook is called
func SubscriptionSuccess(userId string , subscriptionId string, subPaymentId string)bool {
		// update subscription status pending to active

		filter := bson.M{
			"user_id": userId,
			"status":  "pending",
		}
		_, err := connect.SubscriptionCollection.UpdateOne(context.TODO(), filter, bson.D{{Key: "$set", Value: bson.D{{Key: "status", Value: "active"}}}})
		if err != nil {
			fmt.Println("Failed to update subscription status", err)
			return false
		}

		return true
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
			Id:          primitive.NewObjectID().Hex(),
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
func MarkSubscriptionFailed(userId string, subscriptionId string, subPaymentId string) bool {
	
	return false
	
}

func UpdateAutoRenew() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"success": true,
		})
	}
}

type GetActiveSubscriptionResponse struct {
	Plan   string              `json:"plan"`
	Status bool                `json:"status"`
	Data   models.Subscription `json:"data"`
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
					Data:   models.Subscription{},
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
				Data:   subscription,
			})
		}
		return c.Status(fiber.StatusOK).JSON(GetActiveSubscriptionResponse{
			Plan:   subscription.Plan,
			Status: false,
			Data:   subscription,
		})
	}
}


// ************************* Payment webhook ************************


func verifyRazorpaySignature(payload []byte, signature, secret string) bool {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	expected := hex.EncodeToString(h.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}

func PaymentWebhook() fiber.Handler{
	return func (c *fiber.Ctx) error  {
		// payment refund , payment success, failure all must be implemented
		
		envs := env.NewEnv()
		body := c.Body()
		signature  := c.Get("X-RAZORPAY-SIGNATURE")
		secret:= envs.RAZORPAY_WEBHOOK_SECRET

		// verify signature
		if !verifyRazorpaySignature(body, signature, secret){
			return c.SendStatus(fiber.StatusUnauthorized)
		}

		// parse webhook payload
		var event RazorpayEvent
		if err:= json.Unmarshal(body, &event);err!=nil{
			return c.SendStatus(fiber.StatusBadRequest)
		}

		// handle event
		handleRazorpayEvent(event)


		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			
				"message": "successful operation",
			},
		)
	}
}

func handleRazorpayEvent(event RazorpayEvent){
	switch event.Event {
	case "payment.captured":
		p := event.Payload.Payment.Entity
		notes:= p.Notes
		user_id:=notes["user_id"]
		subscription_id := notes["subscription_id"]
		sub_pay_id :=notes["sub_payment_id"]
		SubscriptionSuccess(user_id, subscription_id, sub_pay_id)
		// mark success

		fmt.Println("p:", p)
		
	case "payment.failed":
		// payment failed
		p:= event.Payload.Payment.Entity
		notes:= p.Notes
		user_id:=notes["user_id"]
		subscription_id := notes["subscription_id"]
		sub_pay_id :=notes["sub_payment_id"]
		MarkSubscriptionFailed(user_id, subscription_id, sub_pay_id)
		fmt.Println("p: ", p)
		
	case "refund.processed":
		p:=event.Payload.Payment.Entity
		// update refund status
		fmt.Println("p: ", p)

	default:
		// log unhandled event
		fmt.Println("p")

	}
}

