-- name: GetLaboratories :many
select *
from laboratories
where group_id = coalesce(sqlc.narg(group_id), group_id);

-- name: CreateLaboratory :exec
insert into laboratories (id, group_id, cidr)
values ($1, $2, $3);

-- name: DeleteLaboratory :execrows
delete
from laboratories
where id = $1;