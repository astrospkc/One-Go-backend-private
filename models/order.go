package models

type OrderRequest struct{
	Amount int  `json:"amount"`
	Currency string `json:"currency"`
	// Receipt string `json:"receipt"`
	// PaymentCapture bool `json:"payment_capture"`
}

type OrderResponse struct{
	Id string `json:"id"`
	Amount float64 `json:"amount"`
	Currency string `json:"currency"`
	Status string `json:"status"`
}