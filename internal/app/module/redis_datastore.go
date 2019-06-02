// Copyright 2019 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package module

// RedisDataStore struct
type RedisDataStore struct {
}

// Connect establishes a connection
func (r *RedisDataStore) Connect() (bool, error) {
	return true, nil
}

// Migrate migrates the datastore tables
func (r *RedisDataStore) Migrate() (bool, error) {
	return true, nil
}

// Truncate truncates the datastore tables
func (r *RedisDataStore) Truncate() (bool, error) {
	return true, nil
}

// StoreRelease stores the release data
func (r *RedisDataStore) StoreRelease(release Release) (bool, error) {
	return true, nil
}

// DeleteReleaseByUUID deletes a release data by uuid
func (r *RedisDataStore) DeleteReleaseByUUID(uuid string) (bool, error) {
	return true, nil
}

// DeleteReleaseByURL deletes a release data by url
func (r *RedisDataStore) DeleteReleaseByURL(url string) (bool, error) {
	return true, nil
}

// GetReleaseByUUID gets a release data by uuid
func (r *RedisDataStore) GetReleaseByUUID(uuid string) (*Release, error) {
	return &Release{}, nil
}

// GetReleaseByURL gets a release data by url
func (r *RedisDataStore) GetReleaseByURL(url string) (*Release, error) {
	return &Release{}, nil
}

// GetReleases gets a list of releases
func (r *RedisDataStore) GetReleases(order string) ([]Release, error) {
	return []Release{}, nil
}

// Disconnect closes the connection
func (r *RedisDataStore) Disconnect() (bool, error) {
	return true, nil
}
