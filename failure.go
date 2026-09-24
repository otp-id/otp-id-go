package otpid

// Failure is the delivery failure reason attached to a "failed"
// OrderResult or StatusResult. It is present only when Status == "failed";
// nil for every other status.
type Failure struct {
	// Code is a stable, machine-readable reason. New codes may be added
	// over time — treat any value not in the FailureCode* set as an
	// unspecified failure rather than an error.
	Code string `json:"code"`
	// Message is a human-readable description in Indonesian, suitable for
	// logs but not guaranteed stable across server releases.
	Message string `json:"message"`
}

// Delivery failure codes returned in Failure.Code (kept in sync with the
// server). These describe why an already-accepted transaction failed to
// deliver, which is a distinct contract from the request-level API error
// codes in errors.go (those describe a rejected request and never appear
// here).
const (
	FailureCodeNumberNotOnWhatsApp = "NUMBER_NOT_ON_WHATSAPP"
	FailureCodeTooFrequent         = "TOO_FREQUENT"
	FailureCodeChannelUnavailable  = "CHANNEL_UNAVAILABLE"
	FailureCodeProviderUnavailable = "PROVIDER_UNAVAILABLE"
	FailureCodeDeliveryFailed      = "DELIVERY_FAILED"
)
