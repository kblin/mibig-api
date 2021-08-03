CREATE TABLE IF NOT EXISTS mibig.versions (
    version_id serial PRIMARY KEY,
    name text NOT NULL,
    release_date DATE NOT NULL DEFAULT CURRENT_DATE
);

INSERT INTO mibig.versions (name, release_date)
SELECT val.name, val.release_date::DATE
FROM (
    VALUES
        ('1.0', '2015-06-12'),
        ('1.1', '2015-08-17'),
        ('1.2', '2015-12-24'),
        ('1.3', '2016-09-03'),
        ('1.4', '2018-08-06'),
        ('2.0', '2019-10-16'),
        ('2.1', CURRENT_DATE)
    ) val (name, release_date);


CREATE TABLE IF NOT EXISTS mibig.changelogs (
    changelog_id serial PRIMARY KEY,
    comments text [] NOT NULL,
    version_id int NOT NULL REFERENCES mibig.versions ON DELETE CASCADE,
    entry_id text NOT NULL REFERENCES mibig.entries ON DELETE CASCADE
);
