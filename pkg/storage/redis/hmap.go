package redis

import (
	"encoding/json"
	"github.com/trustwallet/blockatlas/pkg/errors"
	"github.com/trustwallet/blockatlas/pkg/storage/util"
)

func (db *Redis) GetAllHM(entity string) (map[string]string, error) {
	cmd := db.client.HGetAll(entity)
	if cmd.Err() != nil {
		return nil, errors.E(cmd.Err(), util.ErrNotFound)
	}
	return cmd.Val(), nil
}

func (db *Redis) GetHMValue(entity, key string, value interface{}) error {
	cmd := db.client.HMGet(entity, key)
	if cmd.Err() != nil {
		return errors.E(cmd.Err(), util.ErrNotFound)
	}
	val, ok := cmd.Val()[0].(string)
	if !ok {
		return errors.E(util.ErrNotFound)
	}
	err := json.Unmarshal([]byte(val), value)
	if err != nil {
		return errors.E(err, util.ErrNotFound)
	}
	return nil
}

func (db *Redis) AddHM(entity, key string, value interface{}) error {
	j, err := json.Marshal(value)
	if err != nil {
		return errors.E(err, errors.Params{"key": key})
	}
	cmd := db.client.HMSet(entity, map[string]interface{}{key: j})
	if cmd.Err() != nil {
		return errors.E(cmd.Err(), util.ErrNotStored)
	}
	return nil
}

func (db *Redis) DeleteHM(entity, key string) error {
	cmd := db.client.HDel(entity, key)
	if cmd.Err() != nil {
		return errors.E(cmd.Err(), util.ErrNotDeleted)
	}
	return nil
}
