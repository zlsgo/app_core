package service

import (
	"github.com/sohaha/zlsgo/zerror"
	"github.com/sohaha/zlsgo/zlog"
	"github.com/sohaha/zlsgo/zstring"
	"github.com/sohaha/zlsgo/ztime"
	"github.com/sohaha/zlsgo/ztime/cron"
)

type Task struct {
	Run  func()
	Name string
	Cron string
}

// InitTask initializes the tasks using the provided *App.
func InitTask(tasks *[]Task, app *App) (err error) {
	if len(*tasks) == 0 {
		return nil
	}

	t := cron.New()

	app.Log.Debug(app.Log.ColorTextWrap(zlog.ColorLightBlue, zstring.Pad("Cron", 6, " ", zstring.PadLeft)), "Register ")
	for i := range *tasks {
		task := &(*tasks)[i]
		if task.Cron == "" || task.Run == nil {
			continue
		}
		_, err = t.Add(task.Cron, func() {
			err := zerror.TryCatch(func() (err error) {
				task.Run()
				return nil
			})
			if err != nil {
				zlog.Error("Task["+task.Name+"]", err)
				return
			}
		})
		if err != nil {
			return
		}

		next, _ := cron.ParseNextTime(task.Cron)
		app.printLog("", app.Log.ColorTextWrap(zlog.ColorLightGreen, task.Name)+app.Log.ColorTextWrap(zlog.ColorLightWhite, " ["+task.Cron+"]("+ztime.FormatTime(next)+")"))
	}

	t.Run()
	return nil
}
