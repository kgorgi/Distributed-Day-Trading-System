db = db.getSiblingDB("admin");
db.auth("admin", "admin");
db = db.getSiblingDB("audit");

db.createUser({
    user: "user",
    pwd: "user",
    roles: [
        {
            role: "readWrite",
            db: "audit"
        }
    ]
});

db.logs.createIndex({ userID: 1, timestamp: 1, transactionNum: 1 });
