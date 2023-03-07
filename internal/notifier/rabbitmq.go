package notifier

import (
	"context"
	"fmt"
	"github.com/emortalmc/proto-specs/gen/go/message/permission"
	protoModel "github.com/emortalmc/proto-specs/gen/go/model/permission"
	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/protobuf/proto"
	"permission-service-go/internal/config"
	"permission-service-go/internal/repository/model"
	"time"
)

const rabbitMqUriFormat = "amqp://%s:%s@%s:5672"
const exchange = "permission-manager"

type rabbitMqNotifier struct {
	Notifier

	channel *amqp.Channel
}

func NewRabbitMqNotifier(cfg config.RabbitMQConfig) (Notifier, error) {
	conn, err := amqp.Dial(fmt.Sprintf(rabbitMqUriFormat, cfg.Username, cfg.Password, cfg.Host))
	if err != nil {
		return nil, err
	}

	channel, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	return &rabbitMqNotifier{
		channel: channel,
	}, nil
}

func (r *rabbitMqNotifier) RoleUpdate(ctx context.Context, role *model.Role, changeType permission.RoleUpdateMessage_ChangeType) error {
	var protoRole *protoModel.Role
	if role != nil {
		protoRole = role.ToProto()
	}

	msg := permission.RoleUpdateMessage{
		Role:       protoRole,
		ChangeType: changeType,
	}

	bytes, err := proto.Marshal(&msg)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, 5)
	defer cancel()

	return r.channel.PublishWithContext(ctx, exchange, "role_update", false, false, amqp.Publishing{
		ContentType: "application/x-protobuf",
		Timestamp:   time.Now(),
		Type:        string(msg.ProtoReflect().Descriptor().FullName()),
		Body:        bytes,
	})
}

func (r *rabbitMqNotifier) PlayerRolesUpdate(ctx context.Context, playerId string, roleId string, changeType permission.PlayerRolesUpdateMessage_ChangeType) error {
	msg := permission.PlayerRolesUpdateMessage{
		PlayerId:   playerId,
		RoleId:     roleId,
		ChangeType: changeType,
	}

	bytes, err := proto.Marshal(&msg)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, 5)
	defer cancel()

	return r.channel.PublishWithContext(ctx, exchange, "player_role_update", false, false, amqp.Publishing{
		ContentType: "application/x-protobuf",
		Timestamp:   time.Now(),
		Type:        string(msg.ProtoReflect().Descriptor().FullName()),
		Body:        bytes,
	})
}
