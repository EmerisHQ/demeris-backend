package database

import "fmt"

type ErrNoMatchingChannel struct {
	chain_a string
	channel string
	chain_b string
}

func (e ErrNoMatchingChannel) Error() string {
	return fmt.Sprintf("no matching channels found for %s -> %s -> %s", e.chain_a, e.channel, e.chain_b)
}
