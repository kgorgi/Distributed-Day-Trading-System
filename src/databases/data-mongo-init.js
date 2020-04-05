db = db.getSiblingDB("admin");
db.auth("admin", "admin");
db = db.getSiblingDB("extremeworkload");

db.createUser({
    user: "user",
    pwd: "user",
    roles: [
        {
            role: "readWrite",
            db: "extremeworkload"
        }
    ]
});

db.createCollection("users")
db.createCollection("triggers")

db.users.createIndex({ command_id: 1 });
db.triggers.createIndex({ user_command_id: 1, stock: 1, is_sell: 1 });
