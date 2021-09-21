package app

import (
	"context"

	"github.com/reugn/go-quartz/quartz"
)

// StartBackgroundTasks запускает фоновые задачи
func (app *appImpl) StartBackgroundTasks() error {
	app.scheduler.Start()
	app.isSchedulerRunning = true

	// Выгрузка статических данных выполняется каждый день в 9:05 MSK (6:05 UTC)
	err := app.ScheduleBackgroundJob("FetchStaticData", "0 5 6 * * *", func() error {
		return app.FetchStaticData(context.Background())
	})
	if err != nil {
		return err
	}

	// Выгрузка рыночных данных выполняется каждые 15 минут
	err = app.ScheduleBackgroundJob("FetchMarketData", "0 0/15 * * * *", func() error {
		return app.FetchMarketData(context.Background())
	})
	if err != nil {
		return err
	}

	return nil
}

// ScheduleBackgroundJob запускает фоновую задачу
func (app *appImpl) ScheduleBackgroundJob(name, cron string, fn func() error) error {
	trigger, err := quartz.NewCronTrigger(cron)
	if err != nil {
		return err
	}
	job := newBackgroundJob(name, fn)
	err = app.scheduler.ScheduleJob(job, trigger)
	if err != nil {
		return err
	}

	return nil
}

// Close завершает работу приложения
func (app *appImpl) Close() {
	if !app.isSchedulerRunning {
		return
	}

	app.scheduler.Stop()
	app.isSchedulerRunning = true
}

var backgroundJobID = 0

func newBackgroundJob(name string, fn func() error) quartz.Job {
	backgroundJobID++
	id := backgroundJobID

	return &backgroundJob{id, name, fn}
}

type backgroundJob struct {
	id   int
	name string
	fn   func() error
}

func (j *backgroundJob) Execute() {
	err := j.fn()
	if err != nil {
		panic(err)
	}
}

func (j *backgroundJob) Description() string {
	return j.name
}

func (j *backgroundJob) Key() int {
	return j.id
}
