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
package controllers

import (
	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
)

type CronRegistry struct {
	crons     *cron.Cron
	schedules map[string]CronSchedule
}

type CronSchedule struct {
	ID       cron.EntryID
	Schedule string
}

func NewCronRegistry() CronRegistry {
	c := CronRegistry{
		crons:     cron.New(),
		schedules: make(map[string]CronSchedule),
	}

	c.crons.Start()

	return c
}

func (r *CronRegistry) Get(name string) (schedule CronSchedule, ok bool) {
	schedule, ok = r.schedules[name]
	return
}

func (r *CronRegistry) Add(name string, schedule string, cmd func()) error {
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

func (r *CronRegistry) Remove(name string) error {
	schedule, ok := r.Get(name)
	if !ok {
		return errors.New("Cron not found")
	}
	r.crons.Remove(schedule.ID)
	delete(r.schedules, name)
	return nil
}
