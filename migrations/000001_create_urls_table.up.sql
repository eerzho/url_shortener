create table urls (
    id serial primary key,
    short_code varchar(10) unique not null,
    long_url text not null,
    created_at timestamp default now(),
    updated_at timestamp default now()
);
