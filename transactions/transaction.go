package transactions

import "fmt"

type Transaction struct {
	For        string `json:"for"`
	Schmeckles int    `json:"schmeckles"`
	From       string `json:"from"`
	To         string `json:"to"`
}

func (tx Transaction) String() string {
	return fmt.Sprintf("%d schmeckles from %s to %s for \"%s\"", tx.Schmeckles, tx.From, tx.To, tx.For)
}

type Transactions []Transaction

func (txs Transactions) String() string {
	concat := ""

	for _, tx := range txs {
		concat += fmt.Sprintf("%s\n", tx.String())
	}

	return concat
}
