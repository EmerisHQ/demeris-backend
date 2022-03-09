package test_utils

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func RespBodyToMap(jsonReader io.ReadCloser, data *map[string]interface{}, t *testing.T) {
	body, err := ioutil.ReadAll(jsonReader)
	require.NoError(t, err)
	StringToMap(body, data, t)
}

func StringToMap(jsonString []byte, data *map[string]interface{}, t *testing.T) {
	err := json.Unmarshal(jsonString, &data)
	require.NoError(t, err, fmt.Sprintf("tried to unmarshall: %s", jsonString))
}

func RetryOnError(f func() error, interval time.Duration, maxRetries int) error {
	var err error
	for i := 0; i < maxRetries; i++ {
		err = f()
		if err == nil {
			return nil
		}
		time.Sleep(interval)
	}
	return err
}
