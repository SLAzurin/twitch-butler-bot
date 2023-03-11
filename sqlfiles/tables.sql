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
  id SERIAL unique,
  channel_id int null,
  command varchar(40) not null,
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
  cooldown int not null,
  PRIMARY KEY (channel_id, command),
  FOREIGN KEY (channel_id) REFERENCES channels(id) on delete cascade
);
CREATE TABLE IF NOT EXISTS channel_command_perm_overrides (
  id serial primary key,
  channel_command_id int not null,
  username varchar(255) not null,
  allowed boolean not null,
  channel_name varchar(255) not null,
  FOREIGN KEY (channel_command_id) references channel_commands(id) on delete cascade
);
CREATE TABLE IF NOT EXISTS channel_data (
  id varchar(30),
  data json not null,
  channel_id int not null,
  PRIMARY KEY (id, channel_id),
  FOREIGN KEY (channel_id) REFERENCES channels(id) on delete cascade
);
CREATE TABLE IF NOT EXISTS channel_rewards (
  id serial PRIMARY KEY,
  reward_id uuid not null,
  channel_id int not null,
  reward_name varchar(255) not null,
  FOREIGN KEY (channel_id) REFERENCES channels(id) on delete cascade
);
CREATE TABLE IF NOT EXISTS channel_command_aliases (
  id serial PRIMARY KEY,
  alias text not null,
  channel_command_id int not null,
  FOREIGN KEY (channel_command_id) REFERENCES channel_commands(id) on delete cascade
);
/*
 redis cache structure
 autosr
 disabled commands
 autounban users
 
 data loading flow
 Init db connection
 query db for all channels
 fetch channel state from redis, and set those, update those as the app is executed
 
 misc data -> this works: UPDATE channel_data SET data = (data::text::integer + 1)::text::json WHERE channel_id = 2 and id = '!dumpy' RETURNING data;
 */
-- Data to add
INSERT INTO channels
VALUES (1, '#ericarei'),
  (2, '#sangnope'),
  (0, '#settings');
INSERT INTO channel_commands
VALUES (1, 0, '!autosr', true, null, 4, 0),
  (2, 0, '!skip', true, null, 4, 10),
  (3, 2, '!dumpy', true, null, 0, 10),
  (
    4,
    0,
    '!commands',
    false,
    'https://gist.github.com/SLAzurin/f77a54a22bdd0a70ec2d81938d432944',
    0,
    10
  ),
  (
    5,
    0,
    '!azuribot',
    false,
    'desuwa ericareiLurk',
    4,
    10
  ),
  (6, 2, '!sr', true, null, 0, 2),
  (7, 0, '!mr', true, null, 0, 1),
  (8, 0, '!disable', true, null, 4, 0);
INSERT INTO channel_commands VALUES (9, 0, '!allow', true, null, 4, 0);
INSERT INTO channel_commands VALUES (10, 0, '!deny', true, null, 4, 0);
INSERT INTO channel_commands VALUES (11, 0, '!azuriai', true, null, 0, 5);
/* (
 0 = any,
 1 = sub,
 2 = founder,
 3 = vip,
 4 = mod,
 5 = broadcaster
 6 = actualgod
 ) */
INSERT INTO channel_command_aliases (alias, channel_command_id)
VALUES ('!togglesr', 1),
  ('!next', 2),
  ('!help', 4);
INSERT INTO channel_command_perm_overrides
VALUES (
    1,
    6,
    ':stummy!stummy@stummy.tmi.twitch.tv',
    true,
    '#sangnope'
  );
INSERT INTO channel_data
SELECT '!dumpy',
  num::text::json,
  2
FROM sangnope
WHERE id = 1;
INSERT INTO channel_rewards (reward_id, channel_id, reward_name)
VALUES(
    '110b2338-fef9-47c1-be96-39363e0b5c87',
    1,
    'sr_nightbot'
  ),
  (
    '57066ddf-2db9-439f-8a19-561f67c49474',
    2,
    'sr_spotify'
  );
-- select *
-- from channel_commands
--   full outer join channel_command_aliases ON channel_command_aliases.channel_command_id = channel_commands.id
--   left join channels on channels.id = channel_commands.channel_id
--   left join channel_command_perm_overrides ON channel_commands.id = channel_command_perm_overrides.channel_command_id and channel_command_perm_overrides.username = ':stummy!stummy@stummy.tmi.twitch.tv'
--   where (channel_commands.command = '!sr' or channel_command_aliases.alias = '!sr') and (channel_commands.channel_id = 0 or channels.channel_name = '#sangnope');