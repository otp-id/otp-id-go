// Example: WhatsApp Inbound — the USER sends a WhatsApp message to OTP.ID
// instead of typing a code. There is nothing to verify manually: OTP.ID
// matches the incoming message to the transaction automatically. This
// example polls GET /v3/otp/{otp_id} until the status becomes "verified".
//
// Mirrors the docs cURL:
//
//	curl -X POST https://api.otp.id/v3/request \
//	  -d '{"channel": "whatsapp_inbound", "brand": "MyApp", "ttl": 300}'
//
// Do NOT call VerifyOTP for this channel — inbound transactions carry no
// code, so any submission counts as a failed attempt. In production,
// prefer the otp.verified webhook over polling (see the root README).
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	otpid "github.com/otp-id/otp-id-go"
)

func main() {
	apiKey := os.Getenv("OTPID_API_KEY")
	if apiKey == "" {
		log.Fatal("set OTPID_API_KEY first")
	}

	client := otpid.NewClient(apiKey)
	ctx := context.Background()

	res, err := client.RequestOTP(ctx, otpid.OrderParams{
		Channel: otpid.ChannelWhatsAppInbound,
		Brand:   "MyApp",
		TTL:     300,
	})
	if err != nil {
		fatalAPI(err)
	}
	if res.Verification == nil {
		log.Fatal("expected a verification block for whatsapp_inbound")
	}
	fmt.Printf("created: otp_id=%s status=%s\n", res.OtpID, res.Status)
	fmt.Println("ask the user to tap this link and send the pre-filled message:")
	fmt.Println("  " + res.Verification.WaLink)
	fmt.Printf("(or message %q to %s — valid until %s)\n",
		res.Verification.Message, res.Verification.WaNumber, res.Verification.ExpiresAt)

	fmt.Println("waiting for the user's WhatsApp message...")
	for deadline := time.Now().Add(5 * time.Minute); time.Now().Before(deadline); {
		time.Sleep(3 * time.Second)
		st, err := client.OTPStatus(ctx, res.OtpID)
		if err != nil {
			fatalAPI(err)
		}
		if st.Status == "verified" {
			fmt.Println("verified at", st.VerifiedAt)
			return
		}
	}
	fmt.Println("timed out — the user never sent the message")
}

func fatalAPI(err error) {
	var apiErr *otpid.APIError
	if errors.As(err, &apiErr) {
		log.Fatalf("api error %s: %s (http %d)", apiErr.Code, apiErr.Message, apiErr.HTTPStatus)
	}
	log.Fatal(err)
}
