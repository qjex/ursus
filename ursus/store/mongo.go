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

func NewMongoStore(conf *ClientConf) (ProxyStore, error) {
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

func (m *mng) Save(ctx context.Context, proxy Proxy) error {
	filter := bson.D{
		{"port", proxy.Port},
		{"addr", proxy.Addr.String()},
		{"proto", proxy.Proto},
	}
	update := bson.D{{
		"updated",
		proxy.Updated,
	}}
	updateResult, err := m.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return errors.Wrap(err, "error updating saved proxy")
	}
	if updateResult.ModifiedCount == 0 {
		_, err := m.collection.InsertOne(ctx, proxy)
		if err != nil {
			return errors.Wrapf(err, "error inserting new proxy %v", proxy)
		}
	}
	return nil
}

func (m *mng) FindAll(ctx context.Context, page, pageSize int64) ([]Proxy, error) {
	findOptions := options.Find()
	var skip int64
	if page < 1 {
		skip = 0
	} else {
		skip = (page - 1) * pageSize
	}
	findOptions.Skip = &skip
	findOptions.SetLimit(pageSize)
	var results []Proxy

	cur, err := m.collection.Find(ctx, bson.D{{}}, findOptions)
	if err != nil {
		return nil, errors.Wrapf(err, "error getting curse")
	}

	for cur.Next(ctx) {
		var elem Proxy
		err := cur.Decode(&elem)
		if err != nil {
			return nil, errors.Wrapf(err, "error decoding result")
		}
		results = append(results, elem)
	}

	if err := cur.Err(); err != nil {
		return nil, errors.Wrapf(err, "error in cursor")
	}
	_ = cur.Close(ctx)
	return results, nil
}
