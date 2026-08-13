// Example: Missed Call OTP — the code is the last digits of the number
// that calls the user, then verify.
//
// Mirrors the docs cURL:
//
//	curl -X POST https://api.otp.id/v3/request \
//	  -d '{"channel": "misscall", "number": "6281234567890"}'
//
// Notes: brand is not needed (no message body), and otp_length has no
// effect — the code length is set by the telephony vendor. The response's
// verification.prefix is the calling number MINUS the code digits, so the
// UI can render "628559263-____" and ask the user to complete it from
// their missed-call log.
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
		Channel:     otpid.ChannelMisscall,
		Destination: dest,
	})
	if err != nil {
		fatalAPI(err)
	}
	fmt.Printf("calling: otp_id=%s status=%s price=%d\n", res.OtpID, res.Status, res.Price)
	if res.Verification != nil {
		fmt.Printf("the incoming call number starts with: %s (complete the last %d digits)\n",
			res.Verification.Prefix, res.Verification.OtpLength)
	}

	fmt.Print("enter the LAST digits of the number that called: ")
	code, _ := bufio.NewReader(os.Stdin).ReadString('\n')

	v, err := client.VerifyOTP(ctx, res.OtpID, strings.TrimSpace(code))
	if err != nil {
		fatalAPI(err)
	}
	if v.Verified {
		fmt.Println("verified!")
	} else {
		fmt.Println("wrong digits:", v.Reason)
	}
}

func fatalAPI(err error) {
	var apiErr *otpid.APIError
	if errors.As(err, &apiErr) {
		log.Fatalf("api error %s: %s (http %d)", apiErr.Code, apiErr.Message, apiErr.HTTPStatus)
	}
	log.Fatal(err)
}
