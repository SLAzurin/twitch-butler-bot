-- CREATE TABLE IF NOT EXISTS sangnope (
--   id SERIAL PRIMARY KEY,
--   num bigint null,
--   description varchar(255) null,
--   searchkey varchar(255) null
-- );
-- INSERT INTO sangnope
-- VALUES (1, 0, 'dumpy count', '!dumpy');
-- UPDATE sangnope set num = (num + 1) where id = 1 returning num;
-- New db structure
CREATE TABLE IF NOT EXISTS channels (
  id SERIAL PRIMARY KEY,
  channel_name varchar(255) not null
);
CREATE TABLE IF NOT EXISTS channel_commands (
  id SERIAL PRIMARY KEY,
  channel_id int null references channels(id),
  command text not null,
  special boolean not null,
  basic_output text null,
  permission_level int not null,
  /* (
   0 = any,
   1 = sub,
   2 = founder,
   3 = vip,
   4 = mod,
   5 = broadcaster 6 = actualgod
   ) */
  cooldown int not null
);
CREATE TABLE IF NOT EXISTS channel_command_perm_overrides (
  id serial primary key,
  channel_command_id int not null references channel_commands(id),
  username varchar(255) not null,
  allowed boolean not null,
  channel_name varchar(255) not null
);
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
 channel_id nullable
 command
 special T/F
 basic_output nullable
 permission_level (0=any, 1=sub, 2=founder, 3=vip, 4=mod, 5=broadcaster 6=actualgod)
 cooldown (in seconds)
 
 
 channel_command_perm_overrides
 channel_command_id
 user
 allowed (F for deny)
 
 redis cache structure
 
 #channel_state
 autosr
 disabled commands
 autounban users
 
 data loading flow
 Init db connection
 query db for all channels
 fetch channel state from redis, and set those, update those as the app is executed
 
 when command happens it pings db every single time? (yea its fine considering it was that way at LF)
 command timeouts should be kept in ram, not cached, not db.
 permissions should be fetched at the same time with a left join
 
 Thats it
 
 
 1 issue to solve
 misc data -> this works: UPDATE randomtable SET value = (value::bigint + 1)::bytea WHERE id = 1 RETURNING value::bigint;
 
 
 Like this:
 CREATE TABLE channel_data (
 id varchar(30) PRIMARY KEY,
 unstructured_data bytea not null,
 channel_id bigint
 );
 
 
 */