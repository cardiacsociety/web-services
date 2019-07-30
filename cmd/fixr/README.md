# fixr

A general utility for checking and fixing various data issues. 

Performs the following tasks:

1. Sets the short url field in the primary data store (`ol_resource.short_url`), and the corresponding doc in the `Links`
collection, to ensure short link redirection will work.
1. Synchronises the `ol_resource.active` field from primary db with `active` fields in `Resources` and `Links` collections.
1. Removes docs in `Resources` and `Links` collections that have been hard-deleted from primary db.


## Configuration

This utility accesses the data stores directly, so does not require API access.

**Env vars**

```bash
# MySQL connection string
MAPPCPD_MYSQL_URL="dbuser:dbpass@tcp(db.hostname.com:3306)/dbname"

# MongoDB connection string
MAPPCPD_MONGO_URL="mongodb://mongodb.hostname.com/mongodbname"

# MongoDB database name
MAPPCPD_MONGO_DBNAME="mongodbname"

# URL for the short link (linkr) service 
MAPPCPD_SHORT_LINK_URL="https://mapp.to"

# This is a bit of a hack and will be removed at some stage, but is required to 
# prepend the record id in a short link. For example, resource with is 1234 is
# referenced by the short link service as "/r1234". The prefix was put in place
# to distinguish short links for different collections, that may have 
# overlapping id numbers. For now, just stick an "r" here.
MAPPCPD_SHORT_LINK_PREFIX="r"
```

## Flags

`-b` *backdays* - ie, how far back to include records based on `updated_at`, defaults to 1.
`-t` *tasks* to perform, comma-separated list if strings, no default. Options are:
    * `fixResources` - checks and fixes short links, and the active flag for resource records 
    * `pubmedData` - updates `ol_resource.attributes` with additional pubmed info


## Usage

```bash
# check short links for resource updated in the last 24 hours (1 day is default)
$ fixr -t "fixResources"

# check short links for resource updated in the last year
$ fixr -b 365 -t "fixResources"

# update all Pubmed data
$ fixr -b 100000 -t "pubmedData"
```

## Pubmed Rate Limits

Note that from Dec 2018 Pubmed imposed rate limits of 3 requests per second 
without an API Key, or 10 requests per second _with_ an API key. See 
[API Keys section](https://www.ncbi.nlm.nih.gov/books/NBK25497/).

_**A delay of 100 milliseconds has been hard-coded into this script**_ 

 