-- name: ChannelConfig :one
select
coalesce(o.starboard, s.channel_id) as starboard,
coalesce(o.disabled, false) as disabled,
coalesce(o.emoji, s.emoji) as emoji,
coalesce(o.reaction_limit, s.reaction_limit) as reaction_limit,
s.allow_self_star as allow_self_star
from starboard s
left outer join starboard_overrides o on o.channel_id = pggen.arg('channel_id')
where s.guild_id = pggen.arg('guild_id')
limit 1;

-- name: HasOverride :one
select exists(select channel_id from starboard_overrides where channel_id = pggen.arg('channel_id'));

-- name: StarboardMessage :one
select * from starboard_messages
where message_id = pggen.arg('id') or starboard_id = pggen.arg('id');

-- name: AddReaction :exec
insert into starboard_reactions (user_id, message_id) values (pggen.arg('user_id'), pggen.arg('message_id')) on conflict (user_id, message_id) do nothing;

-- name: RemoveReaction :exec
delete from starboard_reactions where user_id = pggen.arg('user_id') and message_id = pggen.arg('message_id');

-- name: ReactionCount :one
select count(*) from starboard_reactions where message_id = pggen.arg('message_id');

-- name: RemoveAllReactions :exec
delete from starboard_reactions where message_id = pggen.arg('message_id');

-- name: RemoveStarboard :exec
delete from starboard_messages where message_id = pggen.arg('message_id')
or starboard_id = pggen.arg('message_id');

-- name: InsertStarboard :one
insert into starboard_messages
(message_id, channel_id, guild_id, starboard_id)
values (
    pggen.arg('message_id'),
    pggen.arg('channel_id'),
    pggen.arg('guild_id'),
    pggen.arg('starboard_id')
) returning *;

-- name: GetStarboard :one
select * from starboard_messages
where message_id = pggen.arg('message_id')
or starboard_id = pggen.arg('message_id');
