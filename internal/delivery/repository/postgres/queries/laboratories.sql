-- name: GetLaboratories :many
select *
from laboratories;

-- name: CreateLaboratory :exec
insert into laboratories (id, cidr)
values ($1, $2);

-- name: DeleteLaboratory :execrows
delete
from laboratories
where id = $1;