package datastore

import (
	"github.com/pkg/errors"
	"gopkg.in/mgo.v2"
)

// MongoDBConnection represents a connection to a MongoDB server
// and includes convenience methods for accessing each collection
type MongoDBConnection struct {
	DSN     string
	DBName  string
	Desc    string
	Session *mgo.Session
}

// MongoQuery is used to map query fields in a request to mgo functions
type MongoQuery struct {
	Find   map[string]interface{} `json:"find"`
	Select map[string]interface{} `json:"select"`
	Limit  int                    `json:"limit"`
	Sort   string                 `json:"sort"`
}

// Do executes the query on a collection (c), and scans the results (r)
func (mq MongoQuery) Do(c *mgo.Collection, r *[]interface{}) error {

	// Start to build the Query...
	q := c.Find(mq.Find).Select(mq.Select).Limit(mq.Limit)

	// ... only add sort if there is a value there
	if len(mq.Sort) > 0 {
		q.Sort(mq.Sort)
	}

	// .All runs the query, scans into r and returns an error, if present
	return q.All(r)
}

// Connect to MongoDB
func (m *MongoDBConnection) Connect() error {
	err := m.checkFields()
	if err != nil {
		return err
	}
	//m.Session, err = mgo.Dial(m.DSN)
        
 dialInfo := mgo.DialInfo{
        Addrs:    []string{
            "devsandbox-shard-00-00.leabq.mongodb.net:27017",
            "devsandbox-shard-00-01.leabq.mongodb.net:27017",
            "devsandbox-shard-00-02.leabq.mongodb.net:27017",
        },
        Username: "intergy-user",
        Password: "189kent$",
    }
    tlsConfig := &tls.Config{}
    dialInfo.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {
        conn, err := tls.Dial("tcp", addr.String(), tlsConfig) // add TLS config
        return conn, err
    }
    m.Session, err = mgo.DialWithInfo(&dialInfo)
	return err
}

// MembersCollection returns a pointer to the Members collection
func (m *MongoDBConnection) MembersCollection() (*mgo.Collection, error) {

	return m.Session.DB(m.DBName).C("Members"), nil
}

// ActivitiesCol returns a pointer to the Activities collection
func (m *MongoDBConnection) ActivitiesCol() (*mgo.Collection, error) {

	return m.Session.DB(m.DBName).C("Activities"), nil
}

// ResourcesCollection returns a pointer to the Resources collection
func (m *MongoDBConnection) ResourcesCollection() (*mgo.Collection, error) {

	return m.Session.DB(m.DBName).C("Resources"), nil
}

// ModulesCollection returns a pointer to the Modules collection
func (m *MongoDBConnection) ModulesCollection() (*mgo.Collection, error) {

	return m.Session.DB(m.DBName).C("Modules"), nil
}

// LinksCol returns a pointer to the Links collection
func (m *MongoDBConnection) LinksCol() (*mgo.Collection, error) {

	return m.Session.DB(m.DBName).C("Links"), nil
}

// RecurringCol returns a pointer to the Recurring collection
func (m *MongoDBConnection) RecurringCol() (*mgo.Collection, error) {

	return m.Session.DB(m.DBName).C("Recurring"), nil
}

// Close terminates the Session
func (m *MongoDBConnection) Close() {
	m.Session.Close()
}

func (m *MongoDBConnection) checkFields() error {
	if m.DSN == "" {
		return errors.New("MongoDBConnection.DSN (data source name / connection string) is not set")
	}
	if m.DBName == "" {
		return errors.New("MongoDBConnection.DBName is not set")
	}
	if m.Desc == "" {
		return errors.New("MongoDBConnection.Desc is not set")
	}
	return nil
}
