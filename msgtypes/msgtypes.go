package msgtypes

import (
	"encoding/json"
	"errors"
	"time"
)

type SetReq struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
	Ttl   Duration    `json:"ttl"`
}

type ErrorResp struct {
	Error string `json:"error"`
}

type ValueResp struct {
	Value interface{} `json:"value"`
}

type KeysResp struct {
	Keys []string `json:"keys"`
}

type Duration time.Duration

func (d *Duration) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case float64:
		*d = Duration(time.Duration(value))
		return nil
	case string:
		tmp, err := time.ParseDuration(value)
		if err != nil {
			return err
		}
		*d = Duration(tmp)
		return nil
	default:
		return errors.New("invalid duration")
	}
}
