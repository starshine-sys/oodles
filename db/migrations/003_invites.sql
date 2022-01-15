-- +migrate Up

-- used for join logging
create table invites (
    code    text    primary key,
    name    text    not null
);

-- used for leave notifications
create table representatives (
    user_id     bigint  primary key,
    description text    not null
);
