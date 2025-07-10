create table clicks (
	id serial primary key,
	url_id integer not null references urls(id),
	ip varchar(45) not null,
	user_agent text not null,
	created_at timestamp default now()
)
