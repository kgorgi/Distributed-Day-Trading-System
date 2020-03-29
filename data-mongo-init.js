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

db.users.createIndex({ command_id: 1 });
db.triggers.createIndex({ user_command_id: 1, stock: 1, is_sell: 1 });
