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
	t := cron.New()
	if len(*tasks) == 0 {
		return nil
	}

	zlog.Debug(zlog.ColorTextWrap(zlog.ColorLightBlue, zstring.Pad("Cron", 6, " ", zstring.PadLeft)), "Register ")
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
		app.printLog("", zlog.ColorTextWrap(zlog.ColorLightGreen, task.Name)+zlog.ColorTextWrap(zlog.ColorLightWhite, " ["+task.Cron+"]("+ztime.FormatTime(next)+")"))
	}

	t.Run()
	return nil
}
