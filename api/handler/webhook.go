package handler

import (
	"encoding/json"
	"fmt"
	"github.com/kermesse-backend/internal/users"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/webhook"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
)

func HandleWebhook(userService users.UsersService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const MaxBodyBytes = int64(65536)
		r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)

		payload, err := readRequestBody(w, r)
		if err != nil {
			return
		}

		event, err := verifyWebhookSignature(payload, r)
		if err != nil {
			http.Error(w, fmt.Sprintf("Webhook signature verification failed: %v", err), http.StatusBadRequest)
			return
		}

		if event.Type == "checkout.session.completed" {
			if err := handleCheckoutSessionCompleted(w, event, userService); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			log.Printf("Unhandled event type: %s", event.Type)
		}

		w.WriteHeader(http.StatusOK)
	}
}

// readRequestBody reads the request body, limits the size, and returns the payload.
func readRequestBody(w http.ResponseWriter, r *http.Request) ([]byte, error) {
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Request Body Read Error", http.StatusServiceUnavailable)
		return nil, err
	}
	return payload, nil
}

// verifyWebhookSignature validates the Stripe webhook signature.
func verifyWebhookSignature(payload []byte, r *http.Request) (stripe.Event, error) {
	signatureHeader := r.Header.Get("Stripe-Signature")
	return webhook.ConstructEvent(payload, signatureHeader, os.Getenv("STRIPE_API_KEY"))
}

// handleCheckoutSessionCompleted handles the "checkout.session.completed" event.
func handleCheckoutSessionCompleted(w http.ResponseWriter, event stripe.Event, usersService users.UsersService) error {
	var session stripe.CheckoutSession
	if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
		return fmt.Errorf("Webhook Error: %v", err)
	}

	balance, userId, err := parseSessionMetadata(session.Metadata)
	if err != nil {
		return fmt.Errorf("Error parsing session metadata: %v", err)
	}

	log.Printf("balance: %v\n", balance)
	log.Printf("userId: %v\n", userId)

	err = usersService.ModifyBalanceFromStripe(userId, balance)
	if err != nil {
		log.Printf("Error updating user balance: %v\n", err)
		return fmt.Errorf("Error updating user balance")
	}

	return nil
}

// parseSessionMetadata parses and validates metadata from the session.
func parseSessionMetadata(metadata map[string]string) (int, int, error) {
	creditStr, ok := metadata["balance"]
	if !ok {
		return 0, 0, fmt.Errorf("Invalid balance")
	}
	credit, err := strconv.Atoi(creditStr)
	if err != nil {
		return 0, 0, fmt.Errorf("Invalid balance value")
	}

	userIdStr, ok := metadata["user_id"]
	if !ok {
		return 0, 0, fmt.Errorf("Invalid user Id")
	}
	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		return 0, 0, fmt.Errorf("Invalid user ID value")
	}

	return credit, userId, nil
}
