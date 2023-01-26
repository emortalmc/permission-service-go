package repository

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"permission-service-go/internal/config"
	"permission-service-go/internal/repository/model"
	"permission-service-go/internal/repository/registrytypes"
	"time"
)

type mongoRepository struct {
	Repository
	database *mongo.Database

	roleCollection   *mongo.Collection
	playerCollection *mongo.Collection
}

var (
	AlreadyHasRoleError = errors.New("player already has role")
	DoesNotHaveRoleError = errors.New("player does not have role")
)

func NewMongoRepository(ctx context.Context, cfg config.MongoDBConfig) (Repository, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.URI).SetRegistry(createCodecRegistry()))
	if err != nil {
		return nil, err
	}

	database := client.Database("permission-service")
	return &mongoRepository{
		database:         database,
		roleCollection:   database.Collection("roles"),
		playerCollection: database.Collection("players"),
	}, nil
}

func (m *mongoRepository) GetRoles(ctx context.Context) ([]*model.Role, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cursor, err := m.roleCollection.Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}

	var mongoResult []model.Role
	err = cursor.All(ctx, &mongoResult)

	slice := make([]*model.Role, len(mongoResult))
	for i := range mongoResult {
		slice[i] = &mongoResult[i]
	}

	return slice, err
}

func (m *mongoRepository) CreateRole(ctx context.Context, role *model.Role) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := m.roleCollection.InsertOne(ctx, role)
	return err
}

func (m *mongoRepository) UpdateRole(ctx context.Context, role *model.Role) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result := m.roleCollection.FindOneAndReplace(ctx, bson.M{"_id": role.Id}, role)
	return result.Err()
}

func (m *mongoRepository) GetPlayerRoleIds(ctx context.Context, playerId uuid.UUID) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	filter := bson.D{{"_id", playerId}}

	var result model.Player
	err := m.playerCollection.FindOne(ctx, filter).Decode(&result)

	if err != nil {
		// insert into db if not exists
		if err == mongo.ErrNoDocuments {
			_, err := m.playerCollection.InsertOne(ctx, model.Player{Id: playerId, Roles: []string{"default"}})

			if err != nil {
				return nil, err
			}

			return []string{"default"}, nil
		} else {
			return nil, err
		}
	}

	return result.Roles, err
}

func (m *mongoRepository) AddRoleToPlayer(ctx context.Context, playerId uuid.UUID, roleId string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := m.playerCollection.UpdateOne(ctx, bson.M{"_id": playerId}, bson.M{"$addToSet": bson.M{"roles": roleId}})

	if err != nil {
		return err
	}

	if result.ModifiedCount == 0 {
		return AlreadyHasRoleError
	}

	return err
}

func (m *mongoRepository) RemoveRoleFromPlayer(ctx context.Context, playerId uuid.UUID, roleId string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := m.playerCollection.UpdateOne(ctx, bson.D{{"_id", playerId}}, bson.M{"$pull": bson.M{"roles": roleId}})

	if err != nil {
		return err
	}
	if result.ModifiedCount == 0 {
		return DoesNotHaveRoleError
	}

	return err
}

func createCodecRegistry() *bsoncodec.Registry {
	return bson.NewRegistryBuilder().
		RegisterTypeEncoder(registrytypes.UUIDType, bsoncodec.ValueEncoderFunc(registrytypes.UuidEncodeValue)).
		RegisterTypeDecoder(registrytypes.UUIDType, bsoncodec.ValueDecoderFunc(registrytypes.UuidDecodeValue)).
		Build()
}
