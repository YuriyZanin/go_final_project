package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

const DATE_FORMAT = "20060102"

type Handler struct {
	service Service
}

func NewHandler(service Service) Handler {
	return Handler{service: service}
}

func (h Handler) getNextDate(w http.ResponseWriter, r *http.Request) {
	nowStr := r.FormValue("now")
	date := r.FormValue("date")
	repeat := r.FormValue("repeat")
	now, err := time.Parse("20060102", nowStr)
	if err != nil {
		http.Error(w, "неверный формат даты", http.StatusBadRequest)
		return
	}

	next, err := h.service.NextDate(now, date, repeat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err = w.Write([]byte(next)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h Handler) postTask(w http.ResponseWriter, r *http.Request) {
	var newTask Task
	var buf bytes.Buffer

	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		http.Error(w, wrappError(err.Error()), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if err = json.Unmarshal(buf.Bytes(), &newTask); err != nil {
		http.Error(w, wrappError(err.Error()), http.StatusBadRequest)
		return
	}

	if newTask.Title == "" {
		http.Error(w, wrappError("не заданы обязательные параметры"), http.StatusBadRequest)
		return
	}

	if newTask.Date == "" {
		newTask.Date = time.Now().Truncate(24 * time.Hour).Format(DATE_FORMAT)
	}

	date, err := time.Parse(DATE_FORMAT, newTask.Date)
	if err != nil {
		http.Error(w, wrappError(err.Error()), http.StatusBadRequest)
		return
	}

	var nextDate string
	if newTask.Repeat == "" {
		nextDate = time.Now().Format(DATE_FORMAT)
	} else {
		nextDate, err = h.service.NextDate(time.Now().Truncate(24*time.Hour), newTask.Date, newTask.Repeat)
		if err != nil {
			http.Error(w, wrappError(err.Error()), http.StatusBadRequest)
			return
		}
	}

	if date.Before(time.Now().Truncate(24 * time.Hour)) {
		newTask.Date = nextDate
	}

	createdTask, err := h.service.Create(newTask)
	if err != nil {
		http.Error(w, wrappError(err.Error()), http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(map[string]any{"id": createdTask.Id})
	if err != nil {
		http.Error(w, wrappError(err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusCreated)
	if _, err = w.Write(resp); err != nil {
		http.Error(w, wrappError(err.Error()), http.StatusInternalServerError)
	}
}

func (h Handler) getTasks(w http.ResponseWriter, r *http.Request) {
	search := r.FormValue("search")
	tasks, err := h.service.FindBy(search)

	if err != nil {
		http.Error(w, wrappError(err.Error()), http.StatusInternalServerError)
		return
	}

	if tasks == nil {
		tasks = make([]Task, 0)
	}

	resp, err := json.Marshal(map[string]any{"tasks": tasks})
	if err != nil {
		http.Error(w, wrappError(err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err = w.Write(resp); err != nil {
		http.Error(w, wrappError(err.Error()), http.StatusInternalServerError)
	}
}

func (h Handler) getTask(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")

	if id == "" {
		http.Error(w, wrappError("Не указан идентификатор"), http.StatusBadRequest)
		return
	}

	n, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, wrappError(err.Error()), http.StatusBadRequest)
		return
	}

	task, err := h.service.FindById(n)
	if err != nil {
		http.Error(w, wrappError(err.Error()), http.StatusBadRequest)
		return
	}

	resp, err := json.Marshal(task)
	if err != nil {
		http.Error(w, wrappError(err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err = w.Write(resp); err != nil {
		http.Error(w, wrappError(err.Error()), http.StatusInternalServerError)
	}
}

func (h Handler) putTask(w http.ResponseWriter, r *http.Request) {
	var task Task
	var buf bytes.Buffer

	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		http.Error(w, wrappError(err.Error()), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if err = json.Unmarshal(buf.Bytes(), &task); err != nil {
		http.Error(w, wrappError(err.Error()), http.StatusInternalServerError)
		return
	}

	if task.Id == "" || task.Title == "" {
		http.Error(w, wrappError("Не указаны необходимые параметры"), http.StatusBadRequest)
		return
	}

	if _, err := strconv.Atoi(task.Id); err != nil {
		http.Error(w, wrappError(err.Error()), http.StatusBadRequest)
		return
	}

	if _, err := time.Parse(DATE_FORMAT, task.Date); err != nil {
		http.Error(w, wrappError(err.Error()), http.StatusBadRequest)
		return
	}

	if _, err := h.service.NextDate(time.Now(), task.Date, task.Repeat); task.Repeat != "" && err != nil {
		http.Error(w, wrappError(err.Error()), http.StatusBadRequest)
		return
	}

	if err := h.service.Update(task); err != nil {
		http.Error(w, wrappError(err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	if _, err = w.Write([]byte("{}")); err != nil {
		http.Error(w, wrappError(err.Error()), http.StatusInternalServerError)
	}
}

func (h Handler) postDone(w http.ResponseWriter, r *http.Request) {
	number := r.FormValue("id")

	id, err := strconv.Atoi(number)
	if err != nil {
		http.Error(w, wrappError(err.Error()), http.StatusBadRequest)
		return
	}

	task, err := h.service.FindById(id)
	if err != nil {
		http.Error(w, wrappError(err.Error()), http.StatusBadRequest)
		return
	}

	if task.Repeat == "" {
		err = h.service.Delete(id)
		if err != nil {
			http.Error(w, wrappError(err.Error()), http.StatusInternalServerError)
			return
		}
	} else {
		nextDate, err := h.service.NextDate(time.Now().Truncate(24*time.Hour), task.Date, task.Repeat)
		if err != nil {
			http.Error(w, wrappError(err.Error()), http.StatusInternalServerError)
			return
		}

		task.Date = nextDate
		if err := h.service.Update(task); err != nil {
			http.Error(w, wrappError(err.Error()), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err = w.Write([]byte("{}")); err != nil {
		http.Error(w, wrappError(err.Error()), http.StatusInternalServerError)
	}
}

func (h Handler) deleteTask(w http.ResponseWriter, r *http.Request) {
	number := r.FormValue("id")

	id, err := strconv.Atoi(number)
	if err != nil {
		http.Error(w, wrappError(err.Error()), http.StatusBadRequest)
		return
	}

	if err = h.service.Delete(id); err != nil {
		http.Error(w, wrappError(err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	if _, err = w.Write([]byte("{}")); err != nil {
		http.Error(w, wrappError(err.Error()), http.StatusInternalServerError)
	}
}

func (h Handler) signin(w http.ResponseWriter, r *http.Request) {
	var buf bytes.Buffer

	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		http.Error(w, wrappError(err.Error()), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var m map[string]string
	if err = json.Unmarshal(buf.Bytes(), &m); err != nil {
		http.Error(w, wrappError(err.Error()), http.StatusBadRequest)
		return
	}

	pass, ok := m["password"]
	if !ok {
		http.Error(w, wrappError("Не задан пароль"), http.StatusBadRequest)
		return
	}

	env := os.Getenv("TODO_PASSWORD")
	var resp map[string]string
	if len(env) > 0 && env == pass {
		resp = map[string]string{"token": hash(env)}
	} else {
		resp = map[string]string{"error": "Неверный пароль"}
	}

	t, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, wrappError(err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err = w.Write(t); err != nil {
		http.Error(w, wrappError(err.Error()), http.StatusInternalServerError)
	}
}

func (h Handler) auth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// смотрим наличие пароля
		pass := os.Getenv("TODO_PASSWORD")
		if len(pass) > 0 {
			var jwt string // JWT-токен из куки
			// получаем куку
			cookie, err := r.Cookie("token")
			if err == nil {
				jwt = cookie.Value
			}
			var valid bool
			// здесь код для валидации и проверки JWT-токена
			// ...
			h := hash(pass)
			valid = h == jwt

			if !valid {
				// возвращаем ошибку авторизации 401
				http.Error(w, "Authentification required", http.StatusUnauthorized)
				return
			}
		}
		next(w, r)
	})
}

func wrappError(message string) string {
	str, _ := json.Marshal(map[string]any{"error": message})
	return string(str)
}

func hash(s string) string {
	h := sha1.New()
	h.Write([]byte(s))
	sha1_hash := hex.EncodeToString(h.Sum(nil))
	return sha1_hash
}

func (h Handler) InitRouter() *chi.Mux {
	r := chi.NewRouter()
	r.Handle("/*", http.FileServer(http.Dir("./web")))
	r.Post("/api/signin", h.signin)
	r.Get("/api/nextdate", h.getNextDate)
	r.Get("/api/tasks", h.auth(h.getTasks))
	r.Get("/api/task", h.auth(h.getTask))
	r.Post("/api/task", h.auth(h.postTask))
	r.Put("/api/task", h.auth(h.putTask))
	r.Delete("/api/task", h.auth(h.deleteTask))
	r.Post("/api/task/done", h.auth(h.postDone))
	return r
}
