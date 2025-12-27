package utils

import (
	"log"
	"time"
)


func LogUserAction(action string, userID int) {
	log.Printf("[AUDIT] Action: %s, UserID: %d, Timestamp: %s",
	action, userID, time.Now().Format(time.RFC3339))
}

func SendNotification(userID int, message string) {
	log.Printf("[NOTIFICATION] UserID: %d, Message: %s, Timestamp: %s",
	userID, message, time.Now().Format(time.RFC3339))
}

func LogError(operation string, err error) {
	log.Printf("[ERROR] Opearation: %s, Error: %v, Timestamp: %s",
	operation, err, time.Now().Format(time.RFC3339))
}