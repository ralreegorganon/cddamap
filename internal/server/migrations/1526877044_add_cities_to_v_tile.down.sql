create or replace view v_tile as
select 
	l.layer_id, 
	case 
		when l.type = 'overmap' then w.name || '/o_' || z || '_tiles' 
		when l.type = 'seen' then w.name || '/' || c.namehash || '_visible_' || z || '_tiles' 
		when l.type = 'seen_solid' then w.name || '/' || c.namehash || '_visible_solid_' || z || '_tiles' 
	end as tile_root
from 
	layer l
	inner join world w
		on w.world_id = l.world_id
	left outer join character c
		on l.character_id = c.character_id