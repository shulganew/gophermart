CREATE TABLE users (
    id SERIAL, 
    user_id UUID NOT NULL UNIQUE, 
    login TEXT NOT NULL UNIQUE, 
    password TEXT NOT NULL
    );

CREATE TYPE processing AS ENUM ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED', 'REGISTERED');

CREATE TABLE orders (
    id SERIAL, 
    user_id UUID NOT NULL REFERENCES users(user_id),
    onumber VARCHAR(20) NOT NULL UNIQUE,
    uploaded TIMESTAMP NOT NULL,
    status processing NOT NULL DEFAULT 'NEW'
    );

CREATE TABLE bonuses (
    id SERIAL, 
    onumber VARCHAR(20) NOT NULL REFERENCES orders(onumber),
    bonus_used NUMERIC DEFAULT 0,
    bonus_accural NUMERIC DEFAULT 0
);



	UPDATE orders SET status = 'PROCESSED' WHERE onumber = '1816868606';
    UPDATE bonuses SET  bonus_accural = '10' WHERE onumber = '1816868606';
1816868606
update ud
set assid = s.assid
from sale s 
where ud.id = s.udid;