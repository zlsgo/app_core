package service

import (
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
func InitTask(tasks *[]Task) (err error) {
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
			task.Run()
		})

		if err != nil {
			return
		}

		next, _ := cron.ParseNextTime(task.Cron)
		PrintLog("", zlog.Log.ColorTextWrap(zlog.ColorLightGreen, task.Name)+zlog.ColorTextWrap(zlog.ColorLightWhite, " ["+task.Cron+"]("+ztime.FormatTime(next)+")"))
	}

	t.Run()
	return nil
}
