package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	database "github.com/sandrinasava/Scheduler/internal/db"
	"github.com/sandrinasava/Scheduler/internal/models"
	_ "modernc.org/sqlite"
)

// oбработчик  для api/task
func TaskHandler(db *sql.DB) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		log.Printf("Получен запрос: %s %s", req.Method, req.URL.Path)

		switch req.Method {

		case http.MethodDelete:
			id := req.FormValue("id")

			// проверяю сущ-е id
			var exists int
			err := db.QueryRow("SELECT 1 FROM scheduler WHERE id = ?", id).Scan(&exists)
			if err != nil {
				SendErrorResponse(res, "записи с указанным id нет", http.StatusBadRequest)
				return
			}
			//удаляю задачу
			_, err = db.Exec("DELETE FROM scheduler WHERE id = ?;", id)
			if err != nil {
				SendErrorResponse(res, "неудачный DELETE запрос", http.StatusBadRequest)
				return
			}
			// если все успешно, отправляю поустой json
			res.Header().Set("Content-Type", "application/json; charset=UTF-8")
			json.NewEncoder(res).Encode(map[string]string{})
			return

		case http.MethodPost:

			if !strings.HasPrefix(req.Header.Get("Content-Type"), "application/json") {
				SendErrorResponse(res, "запрос не содержит json", http.StatusUnsupportedMediaType)
				return
			}

			var task models.Task

			// декод-ю тело запроса
			err := json.NewDecoder(req.Body).Decode(&task)
			if err != nil {
				SendErrorResponse(res, "неудачное декодир-е json", http.StatusBadRequest)
				return
			}

			// ищу новую дату
			date, err := CheckTaskAndFindDate(task)
			if err != nil {
				SendErrorResponse(res, err.Error(), http.StatusBadRequest)
				return
			}
			//добавляю задачу в бд
			ID, err := database.InsertAndReturnID(db, date, task.Title, task.Comment, task.Repeat)
			if err != nil {
				log.Printf("ошибка при добавлении задачи")
				SendErrorResponse(res, "ошибка при добавлении задачи", http.StatusBadRequest)
				return
			}
			res.Header().Set("Content-Type", "application/json; charset=UTF-8")
			response := map[string]interface{}{"id": ID}
			err = json.NewEncoder(res).Encode(response)
			if err != nil {
				SendErrorResponse(res, "Ошибка кодирования в JSON", http.StatusBadRequest)
				return
			}

			return

		case http.MethodGet:
			id := req.FormValue("id")
			if id != "" {
				selectTask := `
		        SELECT * FROM scheduler WHERE id = ?`
				t := models.Task{}
				err := db.QueryRow(selectTask, id).Scan(&t.ID, &t.Date, &t.Title, &t.Comment, &t.Repeat)
				if err != nil {
					SendErrorResponse(res, "неуспешный select запрос", http.StatusBadRequest)
					return
				}
				err = json.NewEncoder(res).Encode(t)
				if err != nil {
					log.Printf("Ошибка кодирования в JSON")
					SendErrorResponse(res, "Ошибка кодирования в JSON", http.StatusBadRequest)
					return

				}
			} else {
				SendErrorResponse(res, "недостаточно параметров", http.StatusMethodNotAllowed)
				return
			}

		case http.MethodPut:
			var task models.Task

			contentType := req.Header.Get("Content-Type")
			log.Println("Content-Type:", contentType)
			if !strings.HasPrefix(req.Header.Get("Content-Type"), "application/json") {
				log.Printf("запрос не содержит json")
				SendErrorResponse(res, "запрос не содержит json", http.StatusUnsupportedMediaType)
				return
			}

			err := json.NewDecoder(req.Body).Decode(&task)
			if err != nil {
				log.Printf("неудачное декодир-е json")
				SendErrorResponse(res, "неудачное декодир-е json", http.StatusBadRequest)
				return
			}

			// ищу новую дату
			date, err := CheckTaskAndFindDate(task)
			if err != nil {
				SendErrorResponse(res, err.Error(), http.StatusBadRequest)
				return
			}

			// проверяю сущ-е id
			var exists int
			err = db.QueryRow("SELECT 1 FROM scheduler WHERE id = ?;", task.ID).Scan(&exists)
			if err != nil {
				SendErrorResponse(res, "записи с указанным id нет", http.StatusBadRequest)
				return
			}
			// изменяю данные в дб
			selectTask := `
	              UPDATE scheduler SET date = $1, title = $2, comment = $3, repeat = $4 where id = $5;
                  `
			_, err = db.Exec(selectTask, date, task.Title, task.Comment, task.Repeat, task.ID)
			if err != nil {
				SendErrorResponse(res, "неудачный update запрос", http.StatusBadRequest)
				return
			}
			// если все успешно, отправляю поустой json
			res.Header().Set("Content-Type", "application/json; charset=UTF-8")
			json.NewEncoder(res).Encode(map[string]string{})
			return

		default:
			SendErrorResponse(res, "неподходящий метод запроса", http.StatusMethodNotAllowed)
			return

		}

	}
}
