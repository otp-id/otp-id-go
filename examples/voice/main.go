// Example: Voice OTP — a phone call speaks a 4-digit code, then verify.
//
// Mirrors the docs cURL:
//
//	curl -X POST https://api.otp.id/v3/request \
//	  -d '{"channel": "voice", "number": "6281234567890", "brand": "MyApp", "ttl": 300}'
//
// Notes: Brand is REQUIRED by the server for voice, and the code length is
// always 4 (any otp_length in the request is ignored for this channel).
package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
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

	client := otpid.NewClient(apiKey)
	ctx := context.Background()

	res, err := client.RequestOTP(ctx, otpid.OrderParams{
		Channel:     otpid.ChannelVoice,
		Destination: dest,
		Brand:       "MyApp", // required for voice
		TTL:         300,
	})
	if err != nil {
		fatalAPI(err)
	}
	fmt.Printf("calling: otp_id=%s status=%s price=%d last_balance=%d\n",
		res.OtpID, res.Status, res.Price, res.LastBalance)

	fmt.Print("enter the 4-digit code spoken in the call: ")
	code, _ := bufio.NewReader(os.Stdin).ReadString('\n')

	v, err := client.VerifyOTP(ctx, res.OtpID, strings.TrimSpace(code))
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
