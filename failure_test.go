package otpid

import "testing"

func TestFailureCodeValues(t *testing.T) {
	// Guards the constants against typos — values are the API contract.
	cases := map[string]string{
		FailureCodeNumberNotOnWhatsApp: "NUMBER_NOT_ON_WHATSAPP",
		FailureCodeTooFrequent:         "TOO_FREQUENT",
		FailureCodeChannelUnavailable:  "CHANNEL_UNAVAILABLE",
		FailureCodeProviderUnavailable: "PROVIDER_UNAVAILABLE",
		FailureCodeDeliveryFailed:      "DELIVERY_FAILED",
	}
	for got, want := range cases {
		if got != want {
			t.Errorf("constant = %q, want %q", got, want)
		}
	}
}
