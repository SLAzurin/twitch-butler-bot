CREATE TABLE IF NOT EXISTS sangnope (
  id SERIAL PRIMARY KEY,
  num bigint null,
  description varchar(255) null,
  searchkey varchar(255) null
);

INSERT INTO sangnope VALUES (1, 0, 'dumpy count', '!dumpy');
/* UPDATE sangnope set num = (num + 1) where id = 1 returning num; */

/*

All states:
disabled commands
autosr
rewardsMap

cache in ram, dont store in db:
lastBanTime
command cooldowns


db structure

channels
  id
  channel

channel_commands
  channel_command_id
  channel_id
  command
  special T/F
  all
  subs
  mods
  super
  

channel_command_perm_overrides
  channel_command_id
  user
  allowed (F for deny)

redis cache structure

#channel_state
  autosr
  disabled commands

data loading flow
  Init db connection
  query db for all channels
  fetch channel state from redis, and set those, update those as the app is executed

  when command happens it pings db every single time? (yea its fine considering it was that way at LF)
    command timeouts should be kept in ram, not cached, not db.
    permissions should be fetched at the same time with a left join
  
  Thats it

*/