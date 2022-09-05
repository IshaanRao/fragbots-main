package database

import (
	"context"
	"fragbotsbackend/constants"
	"fragbotsbackend/logging"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var MongoClient *mongo.Client

func StartClient() {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(constants.MongoURL))
	if err != nil {
		logging.LogFatal("Failed to initialize mongo client")
	}
	if err := client.Ping(context.TODO(), readpref.Primary()); err != nil {
		logging.LogFatal("Failed to ping mongo client, error: " + err.Error())
	}
	logging.Log("Successfully initialized and pinged mongo client")
	MongoClient = client
}

func InsertDocument(collection string, document interface{}) (result *mongo.InsertOneResult, err error) {
	result, err = MongoClient.Database("FragDatabase").Collection(collection).InsertOne(context.TODO(), document)
	return
}

func GetAllDocuments(collection string, filter interface{}, documents interface{}) error {
	result, err := MongoClient.Database("FragDatabase").Collection(collection).Find(context.TODO(), filter)
	if err != nil {
		return err
	}
	err = result.All(context.TODO(), documents)
	if err != nil {
		return err
	}
	return nil
}

func GetDocument(collection string, filter interface{}, document interface{}) error {
	result := MongoClient.Database("FragDatabase").Collection(collection).FindOne(context.TODO(), filter)
	if result.Err() != nil {
		return result.Err()
	}
	err := result.Decode(document)
	if err != nil {
		return err
	}
	return nil
}

func UpdateDocument(collection string, filter interface{}, update interface{}) error {
	return MongoClient.Database("FragDatabase").Collection(collection).FindOneAndUpdate(context.TODO(), filter, bson.D{{"$set", update}}).Err()
}

func UpdateDocumentDelField(collection string, filter interface{}, update interface{}) error {
	return MongoClient.Database("FragDatabase").Collection(collection).FindOneAndUpdate(context.TODO(), filter, bson.D{{"$unset", update}}).Err()
}
func UpdateDocumentIncField(collection string, filter interface{}, update interface{}) error {
	return MongoClient.Database("FragDatabase").Collection(collection).FindOneAndUpdate(context.TODO(), filter, bson.D{{"$inc", update}}).Err()
}

func DeleteDocument(collection string, filter interface{}) bool {
	res := MongoClient.Database("FragDatabase").Collection(collection).FindOneAndDelete(context.TODO(), filter)
	return res.Err() == nil
}

func DocumentExists(collection string, filter interface{}) bool {
	res := MongoClient.Database("FragDatabase").Collection(collection).FindOne(context.TODO(), filter)
	return res.Err() == nil
}
