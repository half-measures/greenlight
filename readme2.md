bin will contain binaries 
cmd/api has main.go, does main server, reading and writing HTTP and auth
internal has API packages, does database interaction
migrations has the SQL migration files for the DB
remote has the config files and setup for eventual linux server
makefile has recipes for common admin tasks like audits, DB migrations

To run, currently we
go run ./cmd/api



## API basics

GET = Used for actions to get info that do not change state of app or any data
POST = used for non-idempotent actions that change state. POST gen used for creating a new resource
PUT = used for idempotent actions that mod resource at URL. PUT generally does replace or update actions
PATCH = Partially updating resource at a URL
DELETE = use for deletin'

So GET /v1/movies/1 would get details of movie?id = 1


## 
| Method | URL Pattern | Handler | Action |
|---|---|---|---|
| GET | /v1/healthcheck | healthcheckHandler | Show App information |
| POST | /v1/movies | CreatemovieHandler | create new movie |
|GET | /v1/movies/:id | showMovieHandler | Show details of a movie | 

-The HTTP Router we are using does not allow Conflicting routes in the API, pat, CHI and mux does. This is bad and good depending on use.

-Responses should be JSON(Strings must be ", not '!)

-json.Decoder will be our pick instead of json.Unmarshal, it needs less code and has more settings we will use.




## Mynotes
When GO encodes a type of JSON, it looks first to see if it had MarshalJSON method on it, if it has, GO calls this method to decide how to encode it. So it uses the json.Marshaler interface to see if it 'satisfies' the interface. 

DSN = Data Source Name, lol

Postgres-------
PGSQL pools have two types of connections
-inUse - doing an actual task
-Idle - 

go will re-use the conn, if none found, will create a new one
bad conn are auto tried twice, if its still bad its auto removed

setmaxopenconns() - sets upper limit of open connections (inuse+idle) Default = unlimted but actual postgres has a 100 HARD limit

SetMaxIdleConns() - Upper idle limit settings, default max is 2
higher num means better perf usually, helps save resources as its not spinning up new connections. but its taking up memory

Once a sec go runs a cleanup to remove expired conn form pools

We should be explicit on MaxOpenConns value to limit an attack and have a built in throttle, we did 25



## SQL used

Fresh install of postgres creates a superuser just calle dpostgres. 

CREATE DATABASE greenlight;
postgres has a little \c meta commond system. 
\l lists all DB
\dt to list all tables
\du to list all users, kidna cool

--Create user + Extension
CREATE ROLE greenlight WITH LOGIN PASSWORD 'pa55word';
CREATE EXTENSION IF NOT EXISTS citext;

can use website PGTUNE to fine tune the config file on the fly
## below is a required export forthe DSN to work
storing as a ENV variable in the machine itsef. Docker may not liek this
export GREENLIGHT_DB_DSN='postgres://greenlight:pa55word@localhost/greenlight'
 
 root =
 sudo -u postgres psql

so had a hell of a time getting permisions,
greenlight user needs to go to the actual DB, and run 
GRANT ALL ON SCHEMA public TO greenlight;

## sql errors wiht migrations
So when you get a error, all SQL will have been executed up to that spot. Its possible the sql was partially applied at that spot. Its best to just drop, and rebuild the table from scratch
