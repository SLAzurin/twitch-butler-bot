CREATE TABLE IF NOT EXISTS sangnope (
  id SERIAL PRIMARY KEY,
  num bigint null,
  description varchar(255) null,
  searchkey varchar(255) null
);

INSERT INTO sangnope VALUES (1, 0, 'dumpy count', '!dumpy');
/* UPDATE sangnope set num = (num + 1) where id = 1 returning num; */