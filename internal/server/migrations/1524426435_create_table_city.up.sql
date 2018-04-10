create table city
(
    city_id serial not null,
    world_id int not null,
    name character varying not null, 
    size int not null, 
    the_geom geometry(POINT) not null,
    created_at timestamp with time zone not null default now(),
    constraint city_pkey primary key (city_id)
);

alter table city add constraint fk_city_world foreign key(world_id) references world(world_id);
create index city_gix ON city using gist (the_geom);

