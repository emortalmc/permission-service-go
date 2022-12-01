package registry

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
)

var (
	MongoRegistry = bson.NewRegistryBuilder().
		RegisterTypeEncoder(UUIDType, bsoncodec.ValueEncoderFunc(UuidEncodeValue)).
		RegisterTypeDecoder(UUIDType, bsoncodec.ValueDecoderFunc(UuidDecodeValue)).
		Build()
)