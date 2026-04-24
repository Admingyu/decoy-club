package types

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

type ObjectID = bson.ObjectID

func NewObjectID() bson.ObjectID {
	return bson.NewObjectID()
}

func ObjectIDFromHex(hex string) (bson.ObjectID, error) {
	return bson.ObjectIDFromHex(hex)
}

func IsZeroObjectID(id bson.ObjectID) bool {
	return id == bson.NilObjectID
}
