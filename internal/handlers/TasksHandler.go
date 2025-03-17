package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	database "github.com/sandrinasava/Scheduler/internal/db"
	"github.com/sandrinasava/Scheduler/internal/models"
	"github.com/sandrinasava/Scheduler/internal/scheduler"
	_ "modernc.org/sqlite"
)

// oбработчик  для api/tasks
func TasksHandler(db *sql.DB) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			SendErrorResponse(res, "неподходящий метод запроса", http.StatusMethodNotAllowed)
			return
		}

		// Инициализирую слайс как пустой слайс
		tasksSlice := []models.Task{}

		search := req.FormValue("search")
		if search != "" {
			log.Printf("search = %s", search)
			D, err := time.Parse("02.01.2006", search)
			if err != nil {
				log.Printf("парсинг даты не удался")
				//если это не дата, ищу соответствие в столбцах title и comment
				selectTask := `
	              SELECT * FROM scheduler
                  WHERE title LIKE '%' || ? || '%'
                  OR comment LIKE '%' || ? || '%'
                  ORDER BY date ASC LIMIT ?;`

				tasksSlice, err = database.FindTasks(db, selectTask, search, search, Limit)
				if err != nil {
					SendErrorResponse(res, err.Error(), http.StatusBadRequest)
					return
				}
				log.Printf("tasksSlice = %+v", tasksSlice)
			} else {
				// ищу по дате
				Dstr := D.Format(scheduler.Format)

				log.Printf("searchDate = %s", Dstr)
				selectTask := `
	            SELECT * FROM scheduler WHERE date LIKE ? ORDER BY date ASC LIMIT ?`
				tasksSlice, err = database.FindTasks(db, selectTask, Dstr, Limit)
				if err != nil {
					SendErrorResponse(res, err.Error(), http.StatusBadRequest)
					return
				}
			}

			// если параметра search нет, ищу все ближайшие задачи
		} else {
			selectTask := `
	         SELECT * FROM scheduler ORDER BY date ASC LIMIT ?`
			var err error
			tasksSlice, err = database.FindTasks(db, selectTask, Limit)
			if err != nil {
				SendErrorResponse(res, err.Error(), http.StatusBadRequest)
				return
			}
		}
		// Структура для ответа
		type TasksResponse struct {
			Tasks []models.Task `json:"tasks"`
		}

		response := TasksResponse{Tasks: tasksSlice}

		err := json.NewEncoder(res).Encode(response)
		if err != nil {
			log.Printf("Ошибка кодирования в JSON")
			SendErrorResponse(res, "Ошибка кодирования в JSON", http.StatusBadRequest)
			return
		}
		return
	}
}
