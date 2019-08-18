package store

import (
	"context"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
)

type mng struct {
	client     *mongo.Client
	collection *mongo.Collection
}

func NewMongoStore(conf *ClientConf) (Store, error) {
	clientOptions := options.Client().ApplyURI("mongodb://" + conf.Host).
		SetAuth(options.Credential{
			AuthMechanism: "SCRAM-SHA-1",
			Username:      conf.User,
			Password:      conf.Passwd,
		})
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return nil, errors.Wrap(err, "error creating mongo connection")
	}
	err = client.Ping(context.TODO(), nil)

	if err != nil {
		return nil, errors.Wrap(err, "error connecting to mongo")
	}
	collection := client.Database("ursus").Collection("proxies")
	return &mng{
		client:     client,
		collection: collection,
	}, nil
}

func (m *mng) Close() {
	err := m.client.Disconnect(context.TODO())
	log.Print(err)
}

func (m *mng) Save(proxy Proxy) error {
	filter := bson.D{
		{"port", proxy.Port},
		{"addr", proxy.Addr.String()},
		{"proto", proxy.Proto},
	}
	update := bson.D{{
		"updated",
		proxy.Updated,
	}}
	updateResult, err := m.collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return errors.Wrap(err, "error updating saved proxy")
	}
	if updateResult.ModifiedCount == 0 {
		_, err := m.collection.InsertOne(context.TODO(), proxy)
		if err != nil {
			return errors.Wrapf(err, "error inserting new proxy %v", proxy)
		}
	}
	return nil
}
