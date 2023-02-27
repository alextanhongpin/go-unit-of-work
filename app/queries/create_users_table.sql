create table if not exists users (
	id uuid default gen_random_uuid(),
	email text not null,

	primary key (id),
	unique (email)
);
