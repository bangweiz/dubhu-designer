package mapper

import "go.mongodb.org/mongo-driver/v2/bson"

// Historical records without a known actor return an empty ID, not an invented user.
func auditID(id bson.ObjectID) string {
	if id.IsZero() {
		return ""
	}

	return id.Hex()
}
