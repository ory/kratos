# SQL Migrations

Migrations consist of one `up` and one `down` file.
To create these SQL migrations, copy the last migration in `./persistence/sql/migrations/sql` and change the timestamp to the current timestamp and the name to the desired name.

If some logic is different for one of the database systems, add the id after the name to the file name.
The content of that file will override the content of the "general" file for that particular DB system.

Example:

`20220802103909000000_courier_send_count.up.sql`
and
`20220802103909000000_courier_send_count.down.sql`

With for example cockroach specific behavior:

`20220802103909000000_courier_send_count.cockroach.up.sql`
and
`20220802103909000000_courier_send_count.cockroach.down.sql`

Replace `cockroach` with `mysql`, `postgres` or `sqlite` if applicable.

## Old Way

To create SQL migrations, target each database individually and run

```
$ dialect=mysql  # or postgres|cockroach|sqlite
$ name=
$ ory dev pop migration create -d=$dialect ./persistence/sql/migrations/templates $name
$ soda generate sql -e mysql -c ./persistence/sql/.soda.yml -p ./persistence/sql/migrations/templates [name]
$ soda generate sql -e sqlite -c ./persistence/sql/.soda.yml -p ./persistence/sql/migrations/templates [name]
$ soda generate sql -e postgres -c ./persistence/sql/.soda.yml -p ./persistence/sql/migrations/templates [name]
$ soda generate sql -e cockroach -c ./persistence/sql/.soda.yml -p ./persistence/sql/migrations/templates [name]
```

and remove the `sqlite` part from the newly generated file to create a SQL migrations that works with all
aforementioned databases.

## Rendering Migrations

Because migrations needs to be backwards compatible, and because fizz migrations might change, we render
fizz migrations to raw SQL statements using `make migrations-render`.

The concrete migrations being applied can be found in [`./migrations/sql`](./migrations/sql).
