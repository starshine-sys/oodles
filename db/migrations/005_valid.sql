-- +migrate Up

create table valid_responses (
    id          serial  primary key,
    response    text    not null,
    user_id     bigint  not null
);
