package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/sandrinasava/Scheduler/internal/models"
	"github.com/sandrinasava/Scheduler/internal/scheduler"
	_ "modernc.org/sqlite"
)

// oбработчик  для api/task/done
func TaskDoneHandler(db *sql.DB) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {

		id := req.FormValue("id")

		// проверяю сущ-е id
		var exists int
		err := db.QueryRow("SELECT 1 FROM scheduler WHERE id = ?;", id).Scan(&exists)
		if err != nil {
			SendErrorResponse(res, "записи с указанным id нет", http.StatusBadRequest)
			return
		}

		var t models.Task
		//нахожу задачу
		err = db.QueryRow("SELECT date, repeat FROM scheduler WHERE id = ?;", id).Scan(&t.Date, &t.Repeat)
		if err != nil {
			SendErrorResponse(res, "неуспешный select запрос", http.StatusBadRequest)
			return
		}
		// если repeat пустой - удаляю задачу
		if t.Repeat == "" {
			_, err := db.Exec("DELETE FROM scheduler WHERE id = ?;", id)
			if err != nil {
				SendErrorResponse(res, "неудачный DELETE запрос", http.StatusBadRequest)
				return
			}
			// если все успешно, отправляю поустой json
			res.Header().Set("Content-Type", "application/json; charset=UTF-8")
			json.NewEncoder(res).Encode(map[string]string{})
			return
		}
		// если repeat есть, ищу новую дату(так как задача уже была в бд, проверок на валидность не делаю)
		nowTime := time.Now()
		now := nowTime.Format(scheduler.Format)
		date, err := scheduler.NextDate(now, t.Date, t.Repeat)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		// выполняю UPDATE запрос
		updateTask := `
			        UPDATE scheduler SET date = ?  where id = ?;
			       `
		_, err = db.Exec(updateTask, date, id)
		if err != nil {
			SendErrorResponse(res, "неудачный UPDATE запрос", http.StatusBadRequest)
			return
		}
		// если все успешно, отправляю поустой json
		res.Header().Set("Content-Type", "application/json; charset=UTF-8")
		json.NewEncoder(res).Encode(map[string]string{})
		return

	}
}
