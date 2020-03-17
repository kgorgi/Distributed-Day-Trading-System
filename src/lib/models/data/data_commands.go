package modelsdata

// UpdateUserCommand should be passed to the data server with the Command "UPDATE_USER"
type UpdateUserCommand struct {
	UserID      string
	Stock       string
	StockAmount int
	Cents       int
}

// ChooseTriggerCommand should be passed to the data server when specifying a trigger
// with the commands "READ_TRIGGER" and "DELETE_TRIGGER"
type ChooseTriggerCommand struct {
	UserID string
	Stock  string
	IsSell bool
}

// UpdateTriggerPriceCommand should be passed to the data server with the
// "UPDATE_TRIGGER_PRICE" command
type UpdateTriggerPriceCommand struct {
	UserID string
	Stock  string
	IsSell bool
	Price  uint64
}

// UpdateTriggerAmountCommand should be passed to the data server with the
// "UPDATE_TRIGGER_AMOUNT" command
type UpdateTriggerAmountCommand struct {
	UserID string
	Stock  string
	IsSell bool
	Amount uint64
}

// PushUserReserveCommand should be passed to the data server with the
// "PUSH_USER_BUY" and "PUSH_USER_SELL"
type PushUserReserveCommand struct {
	UserID   string
	Stock    string
	Cents    uint64
	NumStock uint64
}
