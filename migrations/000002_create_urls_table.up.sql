create table if not exists urls(
    id serial primary key,
    short_code varchar(255) unique not null,
    original_url text not null,
    created_at timestamp default now(),
    updated_at timestamp default now()
);

create trigger update_urls_updated_at
    before update on urls
    for each row execute function update_updated_at_column();
