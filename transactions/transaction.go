package transactions

import "fmt"

type Transaction struct {
	For        string `json:"for"`
	Schmeckles int    `json:"schmeckles"`
	From       string `json:"from"`
	To         string `json:"to"`
}

func (lb Transaction) String() string {
	return fmt.Sprintf("%d schmeckles from %s to %s for \"%s\"", lb.Schmeckles, lb.From, lb.To, lb.For)
}

type Transactions []Transaction

func (txs Transactions) String() string {
	concat := ""

	for _, tx := range txs {
		concat += fmt.Sprintf("%s\n", tx.String())
	}

	return concat
}
