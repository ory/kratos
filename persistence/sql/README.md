# SQL Migrations

To create a new [fizz](https://gobuffalo.io/en/docs/db/fizz/) migration run in the project root:

```
$ name=
$ soda generate fizz $name -c ./contrib/sql/.soda.yml -p ./contrib/sql/migrations
```

To create SQL migrations, target each database individually and run

```
$ soda generate sql -e mysql -c ./contrib/sql/.soda.yml -p ./contrib/sql/migrations [name]
$ soda generate sql -e sqlite -c ./contrib/sql/.soda.yml -p ./contrib/sql/migrations [name]
$ soda generate sql -e postgres -c ./contrib/sql/.soda.yml -p ./contrib/sql/migrations [name]
$ soda generate sql -e cockroach -c ./contrib/sql/.soda.yml -p ./contrib/sql/migrations [name]
```

or, alternative run 

```
$ soda generate sql -c ./contrib/sql/.soda.yml -p ./contrib/sql/migrations 
```

and remove the `sqlite` part from the newly generated file to create a SQL migrations that works with all
aforementioned databases.
