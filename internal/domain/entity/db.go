package entity

type requestIDKey string

// Export variable with private name, only have the key
const RequestIDKey requestIDKey = "requestID"

func (b requestIDKey) UseValue(use string) string {
	if use == string(b) {
		return ""
	}
	return use
}
