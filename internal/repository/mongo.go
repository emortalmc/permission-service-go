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

const (
	databaseName         = "permission-service"
	roleCollectionName   = "roles"
	playerCollectionName = "players"
)

type mongoRepository struct {
	Repository
	database *mongo.Database

	roleCollection   *mongo.Collection
	playerCollection *mongo.Collection
}

var (
	AlreadyHasRoleError  = errors.New("player already has testRole")
	DoesNotHaveRoleError = errors.New("player does not have testRole")
)

func NewMongoRepository(ctx context.Context, cfg config.MongoDBConfig) (Repository, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.URI).SetRegistry(createCodecRegistry()))
	if err != nil {
		return nil, err
	}

	database := client.Database(databaseName)
	return &mongoRepository{
		database:         database,
		roleCollection:   database.Collection(roleCollectionName),
		playerCollection: database.Collection(playerCollectionName),
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

func (m *mongoRepository) GetRole(ctx context.Context, roleId string) (*model.Role, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	filter := bson.M{"_id": roleId}

	var result *model.Role
	err := m.roleCollection.FindOne(ctx, filter).Decode(&result)

	return result, err
}

func (m *mongoRepository) DoesRoleExist(ctx context.Context, roleId string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	filter := bson.M{"_id": roleId}

	count, err := m.roleCollection.CountDocuments(ctx, filter)

	if err != nil {
		return false, err
	}
	return count > 0, nil
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

	filter := bson.M{"_id": playerId}

	var result *model.Player
	err := m.playerCollection.FindOne(ctx, filter).Decode(&result)

	if err != nil {
		// insert into db if not exists
		if err == mongo.ErrNoDocuments {
			_, err := m.playerCollection.InsertOne(ctx, model.Player{Id: playerId, Roles: []string{model.DefaultRoleId}})

			if err != nil {
				return nil, err
			}

			return []string{model.DefaultRoleId}, nil
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
		if result.MatchedCount == 0 {
			// insert into db if not exists
			_, err = m.playerCollection.InsertOne(ctx, model.Player{Id: playerId, Roles: []string{model.DefaultRoleId, roleId}})
			if err != nil {
				return err
			}
			return nil
		}
		return AlreadyHasRoleError
	}

	return err
}

func (m *mongoRepository) RemoveRoleFromPlayer(ctx context.Context, playerId uuid.UUID, roleId string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := m.playerCollection.UpdateOne(ctx, bson.M{"_id": playerId}, bson.M{"$pull": bson.M{"roles": roleId}})

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
