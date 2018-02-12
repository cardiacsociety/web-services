package main

import (
	"flag"
	"fmt"
	"os"
	"time"
	"strconv"
	"strings"

	"github.com/34South/envr"
	"github.com/pkg/errors"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/mappcpd/web-services/internal/resources"
	"github.com/mappcpd/web-services/internal/platform/datastore"
)

type link struct {
	id        int
	title     string
	shortPath string
	shortURL  string
	longURL   string
}

// backdays specifies how far back to include records in whatever task is being performed
var backdays int

// tasksFlag flag is used to specify specific functions to run, comma-separated
var tasksFlag string

var validTasks = []string{"fixResources"}

func init() {
	envr.New("fixrEnv", []string{
		"MAPPCPD_SHORT_LINK_URL",
		"MAPPCPD_SHORT_LINK_PREFIX",
	}).Auto()

	// flags
	flag.IntVar(&backdays, "b", 1, "Specify backdays as an integer > 0")
	flag.StringVar(&tasksFlag, "t", "", "Specify comma-separated list of tasks to run, 'task1, task2, task3'")
}

func main() {

	// Flag check
	flag.Parse()
	if backdays == 1 {
		fmt.Println("Backdays not specified with -b flag, defaulting to 1")
	} else {
		fmt.Println("Checking records updated within the last", backdays, "days")
	}

	if tasksFlag == "" {
		fmt.Println("No tasks specified")
		os.Exit(0)
	}
	t := strings.Replace(tasksFlag, " ", "", -1)
	tasks := strings.Split(t, ",")
	if err := verifyTasks(tasks); err != nil {
		fmt.Println(errors.Cause(err))
		os.Exit(1)
	}

	// connect to data
	datastore.Connect()

	//var err error // don't shadow
	//
	//db, err = sql.Open("mysql", os.Getenv("MAPPCPD_MYSQL_URL"))
	//if err != nil {
	//	log.Fatalln("Could not connect to MySQL server:", os.Getenv("MAPPCPD_MYSQL_URL"), "-", errors.Cause(err))
	//}
	//
	//ddb, err = mgo.Dial(os.Getenv("MAPPCPD_MONGO_URL"))
	//if err != nil {
	//	log.Fatalln("Could not connect to Mongo server:", os.Getenv("MAPPCPD_MONGO_URL"), "-", errors.Cause(err))
	//}

	// run the set of tasks
	for _, v := range tasks {

		if v == "fixResources" {
			fmt.Println("Running task: 'fixResources' ")
			if err := checkShortLinks(); err != nil {
				fmt.Println(errors.Cause(err))
				os.Exit(1)
			}
			if err := syncActiveFlag(); err != nil {
				fmt.Println(errors.Cause(err))
				os.Exit(1)
			}
			fmt.Println("--- done")
		}

		if v == "" {
			// todo fix pubmed data
		}

	}
}

// verifyTasks checks that the task list contains tasks that are valid
func verifyTasks(tasks []string) error {

	for _, t := range tasks {

		f := false
		for _, vt := range validTasks {
			if t == vt {
				f = true
			}
		}

		if f == false {
			return fmt.Errorf("Invalid task: '%s'", t)
		}
	}

	return nil
}

// checkShortLinks checks all the Resource records for a short link, and if incorrect or not found, fixes them.
func checkShortLinks() error {

	// Select resources that start with 'http%' so don't break relative URLs
	// Note the short_url value from the primary record can be NULL.
	// When this is the case the .Scan method below bombs out - hence use of COALESCE.
	query := "SELECT id, name, COALESCE(short_url, ''), resource_url FROM ol_resource " +
		"WHERE `active` = 1 AND `primary` = 1 AND resource_url LIKE 'http%' " +
		"AND updated_at >= NOW() - INTERVAL " + strconv.Itoa(backdays) + " DAY"

	rows, err := datastore.MySQL.Session.Query(query)
	if err != nil {
		return err
	}

	for rows.Next() {
		l := link{}
		err := rows.Scan(&l.id, &l.title, &l.shortURL, &l.longURL)
		if err != nil {
			return err
		}

		// Work out what we expect, or need to set up for a short link
		// Custom short URL based on id of resource with no padding, eg r12, r3435
		l.shortPath = fmt.Sprintf("%v%v", os.Getenv("MAPPCPD_SHORT_LINK_PREFIX"), l.id)
		// The short_url should be
		expectedShortURL := fmt.Sprintf("%v/%v", os.Getenv("MAPPCPD_SHORT_LINK_URL"), l.shortPath)
		// fmt.Printf("/%s -> %s", l.shortPath, expectedShortURL)

		// There are two scenarios:
		// 1. short_url already has the expected value in the primary store
		// Look for any differences between the primary record and the Links doc, if found then sync changes.
		//
		// 2. Missing, invalid or does not match expected based on config
		// Need to set the full short URL in the primary record, and then sync as above

		// Set in primary record, if not the expected value...
		if l.shortURL != expectedShortURL {
			fmt.Printf("/%s -> %s", l.shortPath, expectedShortURL)
			fmt.Println("...no short url - will create one and then sync")
			l.shortURL = expectedShortURL
			err := setShortURL(l.id, l.shortURL)
			if err != nil {
				return err
			}
		}

		// Check for changes and sync if required...
		err = checkSync(l)
		if err != nil {
			return err
		}
	}

	return nil
}

// setShortURL sets the value of the short_url field in the ol_resource (primary) record. It also updates the
// updated_at value to ensure this record will be picked up for sync later on (by mongr)
func setShortURL(id int, shortURL string) error {

	query := `UPDATE ol_resource SET short_url = "%v", updated_at = NOW() WHERE id = %v LIMIT 1`
	query = fmt.Sprintf(query, shortURL, id)
	_, err := datastore.MySQL.Session.Exec(query)

	return errors.Wrap(err, "sql updated failed")
}

// checkSync looks for differences between the primary record and the Links doc, and
// does an upsert of needed. Note that the full short url is only stored in the ol_resource
// record. In the Links doc only the shortPath is stored as all links are relative to whatever
// base URL the short link redirector is using.
func checkSync(l link) error {

	// get link doc from shortPath
	ld, err := getLinkDoc(l.shortPath)
	if err == mgo.ErrNotFound {
		fmt.Println("No Link doc found so need to create one")
	}
	if err != nil && err != mgo.ErrNotFound {
		return errors.Wrap(err, "mongo query failed")
	}

	// No we can check if anything has changed, and decide if we want to
	doSync := false

	// Make sure the link doc has dates set
	nilTime := time.Time{}
	if ld.CreatedAt == nilTime {
		fmt.Println("CreatedAt is nil - will set to now")
		ld.CreatedAt = time.Now()
		doSync = true
	}
	if ld.UpdatedAt == nilTime {
		fmt.Println("UpdatedAt is nil - will set to now")
		ld.UpdatedAt = time.Now()
		doSync = true
	}
	if ld.Title != l.title {
		ld.Title = l.title
		doSync = true
	}
	if ld.LongUrl != l.longURL {
		ld.LongUrl = l.longURL
		doSync = true
	}

	// sync if needed
	if doSync == true {
		fmt.Println("Do the sync!")
		err := sync(ld, l.shortPath)
		return errors.Wrap(err, "sync failed")
	}

	return nil
}

// getLinkDoc fetches a doc from the Links collection.If the doc is not found it
// returns a valid, but empty, value of type resources.Link.
func getLinkDoc(shortPath string) (resources.Link, error) {

	var l resources.Link

	// Always set this as it is used as the KEY for Links docs
	l.ShortUrl = shortPath

	// connect to collection
	c, err := datastore.MongoDB.LinksCol()
	if err != nil {
		return l, errors.Wrap(err, "error connecting to collection")
	}

	// query
	s := bson.M{"shortUrl": shortPath}
	err = c.Find(s).One(&l)
	// Not found is ok, return the empty value
	if err == mgo.ErrNotFound {
		return l, nil /// no error, just not found
	}
	// Some other error
	if err != nil {
		return l, errors.Wrap(err, "mongo query failed")
	}

	// An existing record found
	return l, nil
}

// sync upserts a doc to the Links collection. The selector is the shortPath value, eg /r123.
// This value will always be the same and we don't know if we are creating a new record
// or modifying an existing one. So shortPath is the best identifier.
func sync(ld resources.Link, shortPath string) error {

	c, err := datastore.MongoDB.LinksCol()
	if err != nil {
		return errors.Wrap(err, "error connecting to collection")
	}

	s := bson.M{"shortUrl": shortPath}
	u := bson.M{"$set": ld}
	_, err = c.Upsert(s, u)
	if err != nil {
		return errors.Wrap(err, "upsert failed")
	}

	return nil
}

// getResourceIDs fetches all of the resource ids with the specified active (soft-delete) status, from the primary db.
func getResourceIDs(active bool) ([]int, error) {

	var ids []int

	// active is stored as 0/1 in MySQL, true/false in MongoDB
	var a int
	if active == true {
		a = 1
	} else {
		a = 0
	}

	// Only resources with proper urls, even though there are very few with relative urls
	// Don't need back days - need to check everything.
	query := "SELECT id FROM ol_resource WHERE resource_url LIKE 'http%' AND active = ?"
	rows, err := datastore.MySQL.Session.Query(query, a)
	if err != nil {
		return nil, errors.New("Error executing query - " + err.Error())
	}
	for rows.Next() {
		var id int
		err := rows.Scan(&id)
		if err != nil {
			return nil, errors.New("Error scanning row - " + err.Error())
		}
		ids = append(ids, id)
	}

	return ids, nil
}

// syncActiveFlag ensures that the active status is the same for resources stored in MySQL and in MongoDB.
// The primary DB is the authoritative source so always use the ol_resource.active field value.
func syncActiveFlag() error {

	// Fetch all inactive resources ids from primary db
	inactiveIDs, err := getResourceIDs(false)
	if err != nil {
		return err
	}
	fmt.Println("Found", len(inactiveIDs), "inactive resources")

	// Set the corresponding Resources docs active field to 'false'
	if err := setResourcesActiveField(inactiveIDs, false); err != nil {
		return err
	}
	// Set the corresponding Links docs active field to 'false'
	if err := setLinksActiveField(inactiveIDs, false); err != nil {
		return err
	}

	// Fetch all active resources ids from primary db
	activeIDs, err := getResourceIDs(true)
	if err != nil {
		return err
	}
	fmt.Println("Found", len(activeIDs), "active resources")

	// Set the corresponding Resources docs active field to 'true'
	if err := setResourcesActiveField(activeIDs, true); err != nil {
		return err
	}

	// Set the corresponding Links docs active field to 'false'
	return setLinksActiveField(activeIDs, true)
}

// setResourcesActiveField sets the active field for Resources docs identified by the ids list
func setResourcesActiveField(ids []int, active bool) error {

	if len(ids) == 0 {
		fmt.Println("No ids passed in to setResourcesActiveField()... nothing to do.")
		return nil
	}

	fmt.Printf("Setting %v Resources to active: %v", len(ids), active)

	c, err := datastore.MongoDB.ResourcesCol()
	if err != nil {
		return errors.Wrap(err, "error connecting to collection")
	}

	s := bson.M{"id": bson.M{"$in": ids}}
	u := bson.M{"$set": bson.M{"active": active}}
	ci, err := c.UpdateAll(s, u)
	if err != nil {
		return errors.Wrap(err, "Mongo query error")
	}
	fmt.Println("...", ci.Updated, "records were updated")

	return nil
}

// setLinksActiveField sets the active field for Links docs identified by the ids list. There is no id field
// in Links docs so use the shortUrl by pre-pending 'r', so id 1789 becomes 'r1789'
func setLinksActiveField(ids []int, active bool) error {

	if len(ids) == 0 {
		fmt.Println("No ids passed in to setLinksActiveField()... nothing to do.")
		return nil
	}

	// prepend ids with 'r' to create selector
	var rids []string
	for _, v := range ids {
		rid := os.Getenv("MAPPCPD_SHORT_LINK_PREFIX") + strconv.Itoa(v)
		rids = append(rids, rid)
	}

	fmt.Printf("Setting %v Links to active: %v", len(ids), active)

	c, err := datastore.MongoDB.LinksCol()
	if err != nil {
		return errors.Wrap(err, "error connecting to collection")
	}

	s := bson.M{"shortUrl": bson.M{"$in": rids}}
	u := bson.M{"$set": bson.M{"active": active}}
	ci, err := c.UpdateAll(s, u)
	if err != nil {
		return errors.Wrap(err, "Mongo query error")
	}

	fmt.Println("...", ci.Updated, "records were updated")

	return nil
}
