




// Create an sqobj database with an existing database
db := sqobj.With(db,"schema")

// Register types, set foreign keys
db.Register("file",file{})

// FOREIGN KEY doc.file references file.file
db.Register(doc{},"doc").Reference(file{},"file")

// Create schemas if they don't exist
db.CreateSchema()

