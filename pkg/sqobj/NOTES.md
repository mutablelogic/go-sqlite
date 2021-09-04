

CREATE TABLE IF NOT EXISTS main.file (
    "index" TEXT NOT NULL,
    path TEXT NOT NULL,
    name TEXT NOT NULL,
    is_dir INTEGER NOT NULL,
    ext TEXT,
    modtime TIMESTAMP,
    size INTEGER NOT NULL,
    PRIMARY KEY ("index",path,name)
)

CREATE TABLE IF NOT EXISTS main.filemark (
    "index" TEXT NOT NULL,
    path TEXT NOT NULL,
    name TEXT NOT NULL,
    mark INTEGER NOT NULL,
    idxtime TIMESTAMP,
    PRIMARY KEY ("index",path,name),
    FOREIGN KEY ("index",path,name) REFERENCES file ON DELETE CASCADE
)


-- In the case where there are unique and/or primary key columns:

INSERT INTO main.file (id,index,path,name,is_dir,ext,modtime,size) VALUES (?,?,?,?,?,?,?)
  ON CONFLICT(index,path,name) 
    DO UPDATE SET -- Set non-unique fields
      is_dir=excluded.is_dir, ext=excluded.ext, modtime=excluded.modtime, size=excluded.size 
    WHERE -- Choose the right row
      ("index"=excluded."index" AND path=excluded.path AND name=excluded.name) 
    AND -- Only update when any non-unique fields are different
      (is_dir<>excluded.is_dir OR ext<>excluded.ext OR modtime<>excluded.modtime OR size<>excluded.size)
  ... add additional ON CONFLICT clauses here for each primary key or unique index

Where there is a auto incrementing primary key column, the INSERT statement would need to add
the ID field
