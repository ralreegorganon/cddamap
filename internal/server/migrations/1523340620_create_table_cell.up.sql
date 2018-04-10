create table cell
(
    cell_id serial not null,
    layer_id int not null,
    id character varying,
    name character varying, 
    the_geom geometry(POLYGON) not null,
    created_at timestamp with time zone not null default now(),    
    constraint cell_pkey primary key (cell_id)
);

alter table cell add constraint fk_cell_layer foreign key(layer_id) references layer(layer_id);

create index cell_gix ON cell using gist (the_geom);
create index cell_layer_id on cell (layer_id);