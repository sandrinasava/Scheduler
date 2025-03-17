package handlers

import (
	"fmt"
	"net/http"

	"github.com/sandrinasava/Scheduler/internal/scheduler"
	_ "modernc.org/sqlite"
)

// обработчик для NextDate
func NextDateHandle(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(res, "Ожидается Get запрос", http.StatusMethodNotAllowed)
		return
	}

	now := req.FormValue("now")
	date := req.FormValue("date")
	repeat := req.FormValue("repeat")

	// вызов функции NextDate
	nextDate, err := scheduler.NextDate(now, date, repeat)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	// Успешный ответ
	fmt.Fprintf(res, nextDate)
}
