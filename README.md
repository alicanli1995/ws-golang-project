
# Observer 👁 Server Side Monitoring Tool 

Observer is a server side monitoring tool that allows you to monitor your
application in real time. It is written in Go and uses Pusher to send
real time notifications to the client. 

## 📖 Table of Contents

- [🔨 Build](#-build)
- [📝 Requirements](#-requirements)
- [🚀 Run](#-run)
- [♀ All Flags](#-all-flags)
- [📦 Packages](#-packages)
- [📜 License](#-license)
- [🙏 Acknowledgments](#-acknowledgments)
- [📚 Further Reading](#-further-reading)
- [✍️ Authors](#️-authors)

*You can visualize the data using the [Observer Client](https://github.com/alicanli1995/monitoring-react-project) React project.*

## 🔨 Build

Build in the normal way on Mac/Linux:

~~~
go build -o observer cmd/web/*.go
~~~

Or on Windows:

~~~
go build -o observer.exe cmd/web/.
~~~

Or for a particular platform:

~~~
env GOOS=linux GOARCH=amd64 go build -o observer cmd/web/*.go
~~~

## 📝 Requirements

observer requires:
- Postgres 11 or later (db is set up as a repository, so other databases are possible)
- An account with [Pusher](https://pusher.com/), or a Pusher alternative 
(like [ipê](https://github.com/dimiro1/ipe))

## 🚀 Run

First, make sure ipê is running (if you're using ipê):

On Mac/Linux
~~~
cd ipe
./ipe 
~~~

On Windows
~~~
cd ipe
ipe.exe
~~~

Then, run the infrastructure:
~~~
cd infrastrucutre
docker-compose up -d
~~~

Run with flags:

~~~
./observer \
-dbuser='postgres' \
-pusherHost='localhost' \
-pusherPort='4001' \
-pusherKey='123abc' \
-pusherSecret='abc123' \
-pusherApp="1" \
-pusherSecure=false
~~~~

## ♀ All Flags

~~~~
Usage of ./observer:
  -db string
        database name (default "observer")
  -dbhost string
        database host (default "localhost")
  -dbport string
        database port (default "5432")
  -dbssl string
        database ssl setting (default "disable")
  -dbuser string
        database user
  -domain string
        domain name (e.g. example.com) (default "localhost")
  -identifier string
        unique identifier (default "observer")
  -port string
        port to listen on (default ":4000")
  -production
        application is in production
  -pusherApp string
        pusher app id (default "9")
  -pusherHost string
        pusher host
  -pusherKey string
        pusher key
  -pusherPort string
        pusher port (default "443")
  -pusherSecret string
        pusher secret
   -pusherSecure
        pusher server uses SSL (true or false)
~~~~

## 📦 Packages

- [pq Driver](https://github.com/lib/pq) - PostgreSQL driver for Go
- [Pusher](https://pusher.com/) - APIs to enable devs building realtime features
- [ElasticSearch](https://www.elastic.co/) - Open Source, Distributed, RESTful Search Engine
- [ipê](https://github.com/dimiro1/ipe) - Open source Pusher server implementation compatible with Pusher client libraries written in GO
- [Go-Chi](https://go-chi.io/) - Lightweight, idiomatic and composable router for building Go HTTP services
- [Cron](https://github.com/go-co-op/gocron) - Easy to use, in-process cron libraries
- [JWT](https://github.com/golang-jwt/jwt) - JSON Web Tokens for Go


## 📜 License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details

## 🙏 Acknowledgments

- [Alex Edwards](https://www.alexedwards.net/) - For his amazing book [Let's Go](https://lets-go.alexedwards.net/)
- [Pusher](https://pusher.com/) - For their amazing service

## 📚 Further Reading

- [Let's Go](https://lets-go.alexedwards.net/) - An amazing book by Alex Edwards
- [Go in Action](https://www.manning.com/books/go-in-action) - An amazing book by William Kennedy, Brian Ketelsen & Erik St. Martin
- [Go Web Programming](https://www.manning.com/books/go-web-programming) - An amazing book by Sau Sheong Chang
- [Go Programming Blueprints](https://www.packtpub.com/application-development/go-programming-blueprints-second-edition) - An amazing book by Mat Ryer

## ✍️ Authors

- [Ali CANLI](https://www.linkedin.com/in/ali-canli/)