-- 2021-02-02
-- Add moderation log

-- +migrate Up

create table if not exists mod_log (
    id          serial  primary key,
    guild_id    bigint,

    user_id     bigint  not null,
    mod_id      bigint  not null,

    action_type text    not null,
    reason      text    not null,

    time    timestamp   not null    default (current_timestamp at time zone 'utc'),

    channel_id  bigint,
    message_id  bigint
);
