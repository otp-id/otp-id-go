// Example: SendOTP — you generate the code yourself and OTP.ID only
// delivers it (POST /v3/send). Supported delivery channels: whatsapp,
// sms, email.
//
// Mirrors the docs cURL:
//
//	curl -X POST https://api.otp.id/v3/send \
//	  -d '{"channel": "sms", "number": "6281234567890", "otp": "482913",
//	       "brand": "MyApp", "ttl": 180, "external_id": "order-8822"}'
package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"

	otpid "github.com/otp-id/otp-id-go"
)

func main() {
	apiKey := os.Getenv("OTPID_API_KEY")
	dest := os.Getenv("OTPID_DESTINATION")
	if apiKey == "" || dest == "" {
		log.Fatal("set OTPID_API_KEY and OTPID_DESTINATION first")
	}

	// Generate our own 6-digit code — with SendOTP, code generation and
	// storage are the caller's responsibility; OTP.ID only delivers it.
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		log.Fatal(err)
	}
	code := fmt.Sprintf("%06d", n.Int64())

	client := otpid.NewClient(apiKey)
	ctx := context.Background()

	res, err := client.SendOTP(ctx, code, otpid.OrderParams{
		Channel:     otpid.ChannelSMS,
		Destination: dest,
		Brand:       "MyApp",
		TTL:         180,
		ExternalID:  "order-8822",
	})
	if err != nil {
		fatalAPI(err)
	}
	fmt.Printf("sent our own code: otp_id=%s status=%s price=%d\n", res.OtpID, res.Status, res.Price)

	fmt.Print("enter the code the user received: ")
	entered, _ := bufio.NewReader(os.Stdin).ReadString('\n')

	// Verification still goes through OTP.ID — it stored a hash of the
	// code it delivered, so VerifyOTP works exactly like with RequestOTP.
	v, err := client.VerifyOTP(ctx, res.OtpID, strings.TrimSpace(entered))
	if err != nil {
		fatalAPI(err)
	}
	if v.Verified {
		fmt.Println("verified!")
	} else {
		fmt.Println("wrong code:", v.Reason)
	}
}

func fatalAPI(err error) {
	var apiErr *otpid.APIError
	if errors.As(err, &apiErr) {
		log.Fatalf("api error %s: %s (http %d)", apiErr.Code, apiErr.Message, apiErr.HTTPStatus)
	}
	log.Fatal(err)
}
