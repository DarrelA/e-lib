package entity

type requestIdKey string

// Export variable with private name, only have the key
const RequestIdKey requestIdKey = "requestId"

func (b requestIdKey) UseValue(use string) string {
	if use == string(b) {
		return ""
	}
	return use
}
