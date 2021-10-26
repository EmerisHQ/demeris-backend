package database

import "fmt"

type ErrNoMatchingChannel struct {
	Chain_a string
	Channel string
	Chain_b string
}

func (e ErrNoMatchingChannel) Error() string {
	return fmt.Sprintf("no matching channels found for %s -> %s -> %s", e.Chain_a, e.Channel, e.Chain_b)
}
