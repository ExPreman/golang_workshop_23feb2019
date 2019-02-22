package storage

import (
	"context"

	migration "github.com/payfazz/psql-migration"
)

const applicationID = "d96f0461dc58a3b512752f3de523ba41"

var statements = []string{
	// SQL statement for Version 1
	`
		create table users(
			name text primary key,
			pass text,
			is_admin bool default false,
			counter int default 0
		);

		create table jobs(
			id serial primary key,
			submit timestamp with time zone default now(),
			process timestamp with time zone default null,
			complete timestamp with time zone default null,
			input int default null,
			output int default null
		);

		create function jobs_notify()
			returns trigger
			language plpgsql
		as $$
		declare
			id int;
		begin
			if (tg_op = 'DELETE') then
				id = old."id";
			else
				id = new."id";
			end if;
			perform pg_notify('jobs_notify', id::text || ',' || random()::text);
			return null;
		end $$;

		create trigger jobs_notify_trigger
		after insert or update or delete on jobs
		for each row execute procedure jobs_notify();
	`,

	// SQL statement for Version n
	// ...
}

func (s *Storage) DoMigration(ctx context.Context) error {
	return migration.Migrate(ctx, migration.MigrateParams{
		ErrorLog:      s.errLog,
		Database:      s.db,
		ApplicationID: applicationID,
		Statements:    statements,
	})
}
