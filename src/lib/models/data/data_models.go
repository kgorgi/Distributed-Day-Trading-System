package modelsdata

type Investment struct {
    Stock string `json:"stock" bson:"stock" `
    Amount int `json:"amount" bson:"amount" `
}

type User struct {
    Command_ID string `json:"command_id" bson:"command_id" `
    Cents int `json:"cents" bson:"cents" `
    Investments []Investment `json:"investments" bson:"investments" `
}

type Trigger struct {
    User_Command_ID string `json:"user_command_id" bson:"user_command_id" `
    Stock string `json:"stock" bson:"stock" `
    Price_Cents int `json:"price_cents" bson:"price_cents" `
    Amount_Cents int `json:"amount_cents" bson:"amount_cents" `
    Is_Sell bool `json:"is_sell" bson:"is_sell" `
}