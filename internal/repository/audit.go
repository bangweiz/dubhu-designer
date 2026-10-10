package repository

import (
	"context"
	"fmt"
	"github.com/bangweiz/dubhu-designer/internal/identity"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"strings"
)

// Stamp every write at the persistence boundary using the authenticated actor.
func auditUpdate(ctx context.Context, update any) (any, error) {
	principal, ok := identity.FromContext(ctx)
	if !ok {
		return nil, identity.ErrMissingOrganisation
	}
	if pipeline, ok := update.(mongo.Pipeline); ok {
		result := append(mongo.Pipeline(nil), pipeline...)
		return append(result, bson.D{{Key: "$set", Value: bson.M{"updated_by": principal.AccountID}}}), nil
	}

	source, ok := update.(bson.M)
	if !ok {
		return nil, fmt.Errorf("unsupported audit update")
	}

	result := bson.M{}
	for operator, value := range source {
		result[operator] = value
	}

	set := bson.M{}
	if value, exists := source["$set"]; exists {
		fields, ok := value.(bson.M)
		if !ok {
			return nil, fmt.Errorf("unsupported audit set")
		}

		for key, value := range fields {
			set[key] = value
			if strings.HasSuffix(key, ".updated_at") {
				set[strings.TrimSuffix(key, "updated_at")+"updated_by"] = principal.AccountID
			}
		}
	}

	delete(set, "created_by")
	set["updated_by"] = principal.AccountID
	result["$set"] = set
	return result, nil
}
