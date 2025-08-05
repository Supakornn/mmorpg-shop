package utils

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

func ConvertToObject(id string) bson.ObjectID {
	objectId, _ := bson.ObjectIDFromHex(id)
	return objectId
}
