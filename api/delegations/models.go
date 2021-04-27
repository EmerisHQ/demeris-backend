package delegations

type Delegations struct {
	Delegator   string       `json:"delegator"`
	Delegations []Delegation `json:"delegations"`
}

type Delegation struct {
	ValidatorAddress string `json:"validator_address"`
	Amount           string `json:"amount"`
}
