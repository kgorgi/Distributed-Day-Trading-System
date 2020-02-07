package modelsdata

type Investment struct {
	Stock  string `json:"stock" bson:"stock" `
	Amount uint64 `json:"amount" bson:"amount" `
}

type User struct {
	Command_ID  string       `json:"command_id" bson:"command_id" `
	Cents       uint64       `json:"cents" bson:"cents" `
	Investments []Investment `json:"investments" bson:"investments" `
}

type Trigger struct {
	User_Command_ID string `json:"user_command_id" bson:"user_command_id" `
	Stock           string `json:"stock" bson:"stock" `
	Price_Cents     uint64 `json:"price_cents" bson:"price_cents" `
	Amount_Cents    uint64 `json:"amount_cents" bson:"amount_cents" `
	Is_Sell         bool   `json:"is_sell" bson:"is_sell" `
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
