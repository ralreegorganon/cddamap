create view v_cell as
select 
	w.world_id, l.layer_id, c.cell_id, l.z, c.id, c.name, c.the_geom
from 
	cell c 
	inner join layer l 
		on c.layer_id = l.layer_id
	inner join world w
		on w.world_id = l.world_id;