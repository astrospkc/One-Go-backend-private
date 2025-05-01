package controller

import (
	"crypto/rand"
	"encoding/hex"
)



func GenerateApiKey() (string, error){
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err!=nil{
		return "",nil
	}
	return hex.EncodeToString(bytes), nil
}

