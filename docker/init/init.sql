CREATE USER market WITH ENCRYPTED PASSWORD '1';
CREATE USER praktikum WITH ENCRYPTED PASSWORD 'praktikum';
CREATE DATABASE market;
CREATE DATABASE praktikum;
GRANT ALL PRIVILEGES ON DATABASE market TO market;
GRANT ALL PRIVILEGES ON DATABASE praktikum TO market;

-- need for migrations (issue https://github.com/golang-migrate/migrate/issues/826)
ALTER DATABASE market OWNER TO market;
ALTER DATABASE praktikum OWNER TO market;