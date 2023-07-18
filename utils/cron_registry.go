/*
Copyright 2023 Koor Technologies, Inc. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package utils

import (
	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
)

// This is to make mocking easier
type CronRegistry interface {
	Get(name string) (string, bool)
	Add(name string, schedule string, cmd func()) error
	Remove(name string) error
}

type cronRegistryClient struct {
	crons     *cron.Cron
	schedules map[string]CronSchedule
}

type CronSchedule struct {
	ID       cron.EntryID
	Schedule string
}

func NewCronRegistry() CronRegistry {
	c := &cronRegistryClient{
		crons:     cron.New(),
		schedules: make(map[string]CronSchedule),
	}

	c.crons.Start()

	return c
}

func (r *cronRegistryClient) Get(name string) (string, bool) {
	cs, ok := r.schedules[name]
	return cs.Schedule, ok
}

func (r *cronRegistryClient) Add(name string, schedule string, cmd func()) error {
	id, err := r.crons.AddFunc(schedule, cmd)
	if err != nil {
		return errors.Wrap(err, "Could not parse schedule")
	}

	r.schedules[name] = CronSchedule{
		ID:       id,
		Schedule: schedule,
	}

	return nil
}

func (r *cronRegistryClient) Remove(name string) error {
	cs, ok := r.schedules[name]
	if !ok {
		return errors.New("Cron not found")
	}
	r.crons.Remove(cs.ID)
	delete(r.schedules, name)
	return nil
}
