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