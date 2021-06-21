package datastore

import (
	"log"
	"os"
	"time"

	"github.com/pkg/errors"

	cache "github.com/patrickmn/go-cache"
)

// Datastore contains connections to the various databases
type Datastore struct {
	MySQL   MySQLConnection
	MongoDB MongoDBConnection
	Cache   *cache.Cache
}

// New returns a pointer to a Datastore
func New() *Datastore {

	// Cache is used to store results of 'background' jobs
	//c := cache.New(5*time.Minute, 10*time.Minute)
        c := cache.New(25*time.Minute, 30*time.Minute)

	return &Datastore{
		Cache: c,
	}
}

// ConnectAll establishes sessions with the databases
func (d *Datastore) ConnectAll() error {

	err := d.ConnectMySQL()
	if err != nil {
		return err
	}

	err = d.ConnectMongoDB()
	if err != nil {
		return err
	}

	return nil
}

// ConnectMySQL establishes a Session with the MySQL database
func (d *Datastore) ConnectMySQL() error {

	err := d.MySQL.Connect()
	if err != nil {
		return errors.Wrap(err, "Error connecting to MySQL")
	}
	err = d.MySQL.Session.Ping()
	if err != nil {
		return errors.Wrap(err, "Error communicating with MySQL")
	}

	return nil
}

// ConnectMongoDB establishes a connection to DB2
func (d *Datastore) ConnectMongoDB() error {

	err := d.MongoDB.Connect()
	if err != nil {
		return errors.Wrap(err, "Error connecting to MongoDB")
	}

	err = d.MongoDB.Session.Ping()
	if err != nil {
		log.Fatalf("Error opening MongoDB connection: %s\n", err.Error())
	}

	return nil
}

// FromEnv sets up the default datastore using env vars
func FromEnv() (Datastore, error) {

	DS := New()
	DS.MySQL = MySQLConnection{
		DSN:  os.Getenv("MAPPCPD_MYSQL_URL"),
		Desc: os.Getenv("MAPPCPD_MYSQL_DESC"),
	}
	DS.MongoDB = MongoDBConnection{
		DSN:    os.Getenv("MAPPCPD_MONGO_URL"),
		DBName: os.Getenv("MAPPCPD_MONGO_DBNAME"),
		Desc:   os.Getenv("MAPPCPD_MONGO_DESC"),
	}

	err := DS.ConnectAll()
	return *DS, err
}
