package utils

import (
	"context"
	"os"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DatabaseConnection struct {
	Access     bool
	Uri        string
	Client     *mongo.Client
	Collection *mongo.Collection
}

type WebPage struct {
	Url     string
	Title   string
	Content string
}

func (d *DatabaseConnection) Connect() {
	if d.Access {
		d.Uri = os.Getenv("MONGO_Uri")
		Client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(d.Uri))
		if err != nil {
			panic(err)
		}
		d.Client = Client
		d.Collection = d.Client.Database("webCrawlerArchive").Collection("webPages")
		filter := bson.D{{}}
		d.Collection.DeleteMany(context.TODO(), filter)
	}
}

func (d *DatabaseConnection) Disconnect() {
	if d.Access {
		d.Client.Disconnect(context.TODO())
	}
}

func (d *DatabaseConnection) InsertWebPage(webpage *WebPage) {
	if d.Access {
		d.Collection.InsertOne(context.TODO(), webpage)
	}
}
