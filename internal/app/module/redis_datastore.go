// Copyright 2019 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package module

import (
	"fmt"
	"github.com/clivern/hippo"
	"github.com/clivern/rabbit/internal/app/model"
	"github.com/spf13/viper"
)

// ReleasesURLLookup hash map name
const ReleasesURLLookup = "url_to_uuid_lookup"

// ReleasesUUIDLookup hash map name
const ReleasesUUIDLookup = "uuid_to_url_lookup"

// ReleasesData hash map name
const ReleasesData = "releases_data"

// RedisDataStore struct
type RedisDataStore struct {
	Client *hippo.Redis
}

// Connect establishes a connection
func (r *RedisDataStore) Connect() (bool, error) {
	r.Client = hippo.NewRedisDriver(
		viper.GetString("redis.addr"),
		viper.GetString("redis.password"),
		viper.GetInt("redis.db"),
	)
	return r.Client.Connect()
}

// StoreRelease stores the release data
func (r *RedisDataStore) StoreRelease(release model.Release) (bool, error) {

	value, err := release.ConvertToJSON()

	if err != nil {
		return false, fmt.Errorf("Error while coverting release to json: [%s]", err.Error())
	}

	_, err = r.Client.HSet(
		fmt.Sprintf("%s%s", viper.GetString("database.redis.hash_prefix"), ReleasesData),
		release.UUID,
		value,
	)

	if err != nil {
		return false, fmt.Errorf("Error while storing release data: [%s]", err.Error())
	}

	_, err = r.Client.HSet(
		fmt.Sprintf("%s%s", viper.GetString("database.redis.hash_prefix"), ReleasesURLLookup),
		release.URL,
		release.UUID,
	)

	if err != nil {
		return false, fmt.Errorf("Error while storing release URL: [%s]", err.Error())
	}

	_, err = r.Client.HSet(
		fmt.Sprintf("%s%s", viper.GetString("database.redis.hash_prefix"), ReleasesUUIDLookup),
		release.UUID,
		release.URL,
	)

	if err != nil {
		return false, fmt.Errorf("Error while storing release UUID: [%s]", err.Error())
	}

	return true, nil
}

// DeleteReleaseByUUID deletes a release data by uuid
func (r *RedisDataStore) DeleteReleaseByUUID(uuid string) (bool, error) {
	_, err := r.Client.HDel(
		fmt.Sprintf("%s%s", viper.GetString("database.redis.hash_prefix"), ReleasesUUIDLookup),
		uuid,
	)

	if err != nil {
		return false, err
	}

	return true, nil
}

// DeleteReleaseByURL deletes a release data by url
func (r *RedisDataStore) DeleteReleaseByURL(url string) (bool, error) {
	_, err := r.Client.HDel(
		fmt.Sprintf("%s%s", viper.GetString("database.redis.hash_prefix"), ReleasesURLLookup),
		url,
	)

	if err != nil {
		return false, err
	}

	return true, nil
}

// GetReleaseByUUID gets a release data by uuid
func (r *RedisDataStore) GetReleaseByUUID(uuid string) (*model.Release, error) {
	return &model.Release{}, nil
}

// GetReleaseByURL gets a release data by url
func (r *RedisDataStore) GetReleaseByURL(url string) (*model.Release, error) {
	return &model.Release{}, nil
}

// GetReleases gets a list of releases
func (r *RedisDataStore) GetReleases(order string) ([]model.Release, error) {
	return []model.Release{}, nil
}
