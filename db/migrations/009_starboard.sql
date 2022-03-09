-- 2021-03-09
-- Add starboard

-- +migrate Up

create table starboard (
    guild_id        bigint  primary key,
    channel_id      bigint  not null default 0,
    emoji           text    not null default '⭐',
    reaction_limit  int     not null default 3,
    allow_self_star bool    not null default true
);

-- inherits from the global starboard table if a value is null
create table starboard_overrides (
    channel_id      bigint  primary key, -- channel or category
    guild_id        bigint  not null,
    disabled        boolean not null default false,
    starboard       bigint,
    emoji           text,
    reaction_limit  int
);

create table starboard_messages (
    message_id      bigint  primary key,
    channel_id      bigint  not null,
    guild_id        bigint  not null,
    starboard_id    bigint  not null
);

create table starboard_reactions (
    user_id     bigint  not null,
    message_id  bigint  not null,

    primary key (user_id, message_id)
);

create index starboard_overrides_guild_idx on starboard_overrides (guild_id);
create index starboard_reactions_message_idx on starboard_reactions (message_id);
