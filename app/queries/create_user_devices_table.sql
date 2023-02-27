create table if not exists user_devices (
	id uuid default gen_random_uuid(),
	user_id uuid not null,
	device_id text not null,

	primary key (id),
	foreign key (user_id) references users(id),
	unique (device_id)
);
