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

// StoreProject stores the project data
func (r *RedisDataStore) StoreProject(project *model.Project) (bool, error) {
	jsonValue, err := project.ConvertToJSON()

	if err != nil {
		return false, fmt.Errorf("Error while coverting project to json: [%s]", err.Error())
	}

	_, err = r.Client.HSet(
		fmt.Sprintf("%s%s", viper.GetString("database.redis.hash_prefix"), ReleasesData),
		project.UUID,
		jsonValue,
	)

	if err != nil {
		return false, fmt.Errorf("Error while storing project data: [%s]", err.Error())
	}

	_, err = r.Client.HSet(
		fmt.Sprintf("%s%s", viper.GetString("database.redis.hash_prefix"), ReleasesURLLookup),
		project.URL,
		project.UUID,
	)

	if err != nil {
		return false, fmt.Errorf("Error while storing project URL: [%s]", err.Error())
	}

	_, err = r.Client.HSet(
		fmt.Sprintf("%s%s", viper.GetString("database.redis.hash_prefix"), ReleasesUUIDLookup),
		project.UUID,
		project.URL,
	)

	if err != nil {
		return false, fmt.Errorf("Error while storing project UUID: [%s]", err.Error())
	}

	return true, nil
}

// DeleteProjectByUUID deletes a project data by uuid
func (r *RedisDataStore) DeleteProjectByUUID(uuid string) (bool, error) {
	_, err := r.Client.HDel(
		fmt.Sprintf("%s%s", viper.GetString("database.redis.hash_prefix"), ReleasesUUIDLookup),
		uuid,
	)

	if err != nil {
		return false, err
	}

	return true, nil
}

// DeleteProjectByURL deletes a project data by url
func (r *RedisDataStore) DeleteProjectByURL(url string) (bool, error) {
	_, err := r.Client.HDel(
		fmt.Sprintf("%s%s", viper.GetString("database.redis.hash_prefix"), ReleasesURLLookup),
		url,
	)

	if err != nil {
		return false, err
	}

	return true, nil
}

// ProjectExistsByURL check if project exists by URL
func (r *RedisDataStore) ProjectExistsByURL(url string) (bool, error) {
	return r.Client.HExists(
		fmt.Sprintf("%s%s", viper.GetString("database.redis.hash_prefix"), ReleasesURLLookup),
		url,
	)
}

// ProjectExistsByUUID check if project exists by UUID
func (r *RedisDataStore) ProjectExistsByUUID(uuid string) (bool, error) {
	return r.Client.HExists(
		fmt.Sprintf("%s%s", viper.GetString("database.redis.hash_prefix"), ReleasesUUIDLookup),
		uuid,
	)
}

// GetProjectByUUID gets a project data by uuid
func (r *RedisDataStore) GetProjectByUUID(uuid string) (*model.Project, error) {
	result, err := r.Client.HGet(
		fmt.Sprintf("%s%s", viper.GetString("database.redis.hash_prefix"), ReleasesData),
		uuid,
	)

	if err != nil {
		return nil, err
	}

	project := model.NewProject()

	project.LoadFromJSON([]byte(result))

	return project, nil
}

// GetProjectByURL gets a project data by url
func (r *RedisDataStore) GetProjectByURL(url string) (*model.Project, error) {
	uuid, err := r.Client.HGet(
		fmt.Sprintf("%s%s", viper.GetString("database.redis.hash_prefix"), ReleasesURLLookup),
		url,
	)

	if err != nil {
		return nil, err
	}

	return r.GetProjectByUUID(uuid)
}

// GetProjects gets projects list
func (r *RedisDataStore) GetProjects() ([]*model.Project, error) {
	var projects []*model.Project

	iter := r.Client.HScan(
		fmt.Sprintf("%s%s", viper.GetString("database.redis.hash_prefix"), ReleasesData),
		0,
		"",
		0,
	).Iterator()

	for iter.Next() {
		data := iter.Val()

		project := model.NewProject()

		_, err := project.LoadFromJSON([]byte(data))

		if err != nil {
			continue
		}

		projects = append(projects, project)
	}

	return projects, nil
}
