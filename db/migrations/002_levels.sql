-- +migrate Up

create table level_config (
    id  bigint  primary key,

    blocked_channels    bigint[]    not null    default array[]::bigint[],
    blocked_roles       bigint[]    not null    default array[]::bigint[],
    blocked_categories  bigint[]    not null    default array[]::bigint[],

    between_xp      interval    not null    default '1 minute',
    reward_text     text        not null    default '',
    levels_enabled  boolean     not null    default true,

    reward_log      bigint  not null    default 0,
    nolevels_log    bigint  not null    default 0,

    dm_on_reward    boolean not null    default false
);

create table level_backgrounds (
    id      serial  primary key,
    name    text    not null,
    source  text    not null,
    blob    bytea   not null,

    emoji_name  text    not null,
    emoji_id    bigint
);

create table levels (
    guild_id    bigint  not null,
    user_id     bigint  not null,

    xp          bigint  not null    default 0,
    colour      bigint  not null    default 0,
    background  bigint              references level_backgrounds (id) on delete set null,

    last_xp timestamp   not null    default (current_timestamp at time zone 'utc'),

    primary key (guild_id, user_id)
);

create table level_rewards (
    guild_id    bigint  not null,
    lvl         bigint  not null,
    role_reward bigint  not null,

    primary key (guild_id, lvl)
);

create table nolevels (
    guild_id    bigint  not null,
    user_id     bigint  not null,

    expires boolean     not null    default false,
    -- this default is not used if "expires" is also left as the default so it's fine
    expiry  timestamp   not null    default (current_timestamp at time zone 'utc'),

    primary key (guild_id, user_id)
);
