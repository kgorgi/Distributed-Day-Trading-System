package modelsdata

type Investment struct {
	Stock  string `json:"stock" bson:"stock" `
	Amount uint64 `json:"amount" bson:"amount" `
}

type User struct {
	Command_ID  string       `json:"command_id" bson:"command_id" `
	Cents       uint64       `json:"cents" bson:"cents" `
	Investments []Investment `json:"investments" bson:"investments" `
	Buys        []Reserve    `json:"buys" bson:"buys"`
	Sells       []Reserve    `json:"sells" bson:"sells"`
}

type Reserve struct {
	Stock      string `json:"stock" bson:"stock"`
	Cents      uint64 `json:"amount" bson:"amount" `
	Num_Stocks uint64 `json:"num_stocks" bson:"num_stocks"`
	Timestamp  uint64 `json:"timestamp" bson:"timestamp"`
}

type Trigger struct {
	User_Command_ID    string `json:"user_command_id" bson:"user_command_id" `
	Stock              string `json:"stock" bson:"stock" `
	Price_Cents        uint64 `json:"price_cents" bson:"price_cents" `
	Amount_Cents       uint64 `json:"amount_cents" bson:"amount_cents" `
	Is_Sell            bool   `json:"is_sell" bson:"is_sell" `
	Transaction_Number uint64 `json:"transaction_number" bson:"transaction_number" `
}

type TriggerDisplayInfo struct {
	Stock        string `json:"stock" bson:"stock" `
	Price_Cents  uint64 `json:"price_cents" bson:"price_cents" `
	Amount_Cents uint64 `json:"amount_cents" bson:"amount_cents" `
	Is_Sell      bool   `json:"is_sell" bson:"is_sell" `
}

type UserDisplayInfo struct {
	Cents       uint64               `json:"cents" bson:"cents" `
	Investments []Investment         `json:"investments" bson:"investments" `
	Triggers    []TriggerDisplayInfo `json:"triggers" bson:"triggers" `
}
