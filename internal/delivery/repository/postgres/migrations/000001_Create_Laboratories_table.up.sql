create table if not exists laboratories
(
    id         uuid        not null primary key,
    group_id   uuid        not null,

    cidr       cidr        not null,

    updated_at timestamptz,

    created_at timestamptz not null default now()
);