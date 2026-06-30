package app

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gen2brain/beeep"
	"github.com/linfree/open-todo/internal/store"
)

type ReminderItem struct {
	Date   string `json:"date"`
	Repeat string `json:"repeat"`
}

func StartReminderChecker(st *store.Store, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			checkReminders(st)
		}
	}()
}

func checkReminders(st *store.Store) {
	tasks, err := st.GetTasks()
	if err != nil {
		log.Printf("[reminder] get tasks: %v", err)
		return
	}

	now := time.Now().Unix()

	for _, task := range tasks {
		if task.Completed || task.Deleted {
			continue
		}

		var reminders []ReminderItem
		if err := json.Unmarshal([]byte(task.Reminders), &reminders); err != nil {
			continue
		}

		for _, r := range reminders {
			t, err := time.Parse(time.RFC3339, r.Date)
			if err != nil {
				continue
			}
			reminderTime := t.Unix()

			if reminderTime > now {
				continue
			}

			sent, err := st.IsReminderSent(task.ID, reminderTime)
			if err != nil || sent {
				continue
			}

			if err := beeep.Notify("Open Todo - 任务提醒", task.Title, ""); err != nil {
				log.Printf("[reminder] notify: %v", err)
			}

			data, _ := json.Marshal(r)
			st.MarkReminderSent("", task.ID, reminderTime, string(data))
		}
	}
}
