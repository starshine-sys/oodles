-- +migrate Up

create table messages (
    id          bigint  primary key,
    user_id     bigint  not null    default 0,
    channel_id  bigint  not null    default 0,
    server_id   bigint  not null    default 0,

    username    text    not null default '',
    member      text,
    system      text,

    content     text    not null
);
