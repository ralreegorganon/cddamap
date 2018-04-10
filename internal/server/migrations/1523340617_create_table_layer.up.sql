create table layer
(
    layer_id serial not null,
    world_id int not null,
    character_id int null,
    z int not null,
    type character varying not null, 
    created_at timestamp with time zone not null default now(),    
    constraint layer_pkey primary key (layer_id)
);

alter table layer add constraint fk_layer_world foreign key(world_id) references world(world_id);
alter table layer add constraint fk_layer_character foreign key(character_id) references character(character_id);
