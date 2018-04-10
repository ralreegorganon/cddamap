create table world
(
    world_id serial not null,
    name character varying not null unique, 
    maxz int not null, 
    created_at timestamp with time zone not null default now(),
    constraint world_pkey primary key (world_id)
);