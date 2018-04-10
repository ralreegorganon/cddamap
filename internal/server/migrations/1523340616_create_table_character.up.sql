create table character
(
    character_id serial not null,
    world_id int not null,
    namehash character varying not null, 
    name character varying not null, 
    created_at timestamp with time zone not null default now(),    
    constraint character_pkey primary key (character_id)
);

alter table character add constraint fk_character_world foreign key(world_id) references world(world_id);
