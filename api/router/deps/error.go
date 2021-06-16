package deps

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/sony/sonyflake"
)

var flake *sonyflake.Sonyflake
var once sync.Once

func init() {
	once.Do(func() {
		flake = sonyflake.NewSonyflake(sonyflake.Settings{})
	})
}

type Error struct {
	ID            string `json:"id"`
	Namespace     string `json:"namespace"`
	StatusCode    int    `json:"-"`
	LowLevelError error  `json:"-"`
	Cause         string `json:"cause"`
}

func (e Error) Error() string {
	return fmt.Sprintf("%s @ %s: %s", e.ID, e.Namespace, e.LowLevelError.Error())
}

func (e Error) Unwrap() error {
	return e.LowLevelError
}

func NewError(namespace string, cause error, statusCode int) Error {
	id, err := flake.NextID()
	if err != nil {
		panic(fmt.Errorf("cannot create sonyflake, %w", err))
	}

	idstr := strconv.FormatUint(id, 10)

	return Error{
		ID:            idstr,
		StatusCode:    statusCode,
		Namespace:     namespace,
		LowLevelError: cause,
		Cause:         cause.Error(),
	}
}
