// Copyright 2019 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package module

// DataStore struct
type DataStore struct {
}

// Migrate migrates the datastore tables
func (d *DataStore) Migrate() (bool, error) {
	return true, nil
}

// Truncate truncates the datastore tables
func (d *DataStore) Truncate() (bool, error) {
	return true, nil
}

// StoreRelease stores the release data
func (d *DataStore) StoreRelease(release Release) (bool, error) {
	return true, nil
}

// DeleteReleaseByUUID deletes a release data by uuid
func (d *DataStore) DeleteReleaseByUUID(uuid string) (bool, error) {
	return true, nil
}

// DeleteReleaseByURL deletes a release data by url
func (d *DataStore) DeleteReleaseByURL(url string) (bool, error) {
	return true, nil
}

// GetReleaseByUUID gets a release data by uuid
func (d *DataStore) GetReleaseByUUID(uuid string) (*Release, error) {
	return &Release{}, nil
}

// GetReleaseByURL gets a release data by url
func (d *DataStore) GetReleaseByURL(url string) (*Release, error) {
	return &Release{}, nil
}

// GetReleases gets a list of releases
func (d *DataStore) GetReleases(order string) ([]Release, error) {
	return []Release{}, nil
}
