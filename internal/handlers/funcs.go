package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/sandrinasava/Scheduler/internal/models"
	"github.com/sandrinasava/Scheduler/internal/scheduler"
	_ "modernc.org/sqlite"
)

const Limit = "15"

func SendErrorResponse(res http.ResponseWriter, message string, statusCode int) {
	res.Header().Set("Content-Type", "application/json; charset=UTF-8")
	res.WriteHeader(statusCode)
	json.NewEncoder(res).Encode(map[string]string{"error": message})
}

func CheckTaskAndFindDate(task models.Task) (string, error) {

	if task.Title == "" {
		return "", fmt.Errorf("параметр title пустой")
	}

	var date time.Time
	if task.Date != "" {
		date, err := time.Parse(scheduler.Format, task.Date)
		if err != nil {
			return "", fmt.Errorf("указан неверный формат даты %v", err)
		}
		if date.Format(scheduler.Format) < time.Now().Format(scheduler.Format) {

			if task.Repeat != "" {

				now := time.Now().Format(scheduler.Format)
				dateStr, err := scheduler.NextDate(now, task.Date, task.Repeat)
				if err != nil {
					return "", fmt.Errorf("ошибка поиска NextDate %v", err)
				}

				log.Printf("следующая дата = %s", dateStr)
				return dateStr, nil

			} else {
				date = time.Now()
				return date.Format(scheduler.Format), nil

			}
		}
		return date.Format(scheduler.Format), nil
	}

	date = time.Now()
	return date.Format(scheduler.Format), nil

}
