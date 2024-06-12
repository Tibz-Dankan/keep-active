package request

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/Tibz-Dankan/keep-active/internal/event"
	"github.com/Tibz-Dankan/keep-active/internal/middlewares"
	"github.com/Tibz-Dankan/keep-active/internal/models"
	"github.com/Tibz-Dankan/keep-active/internal/services"
	"github.com/gorilla/mux"
)

func sendMessage(message, userId string, clientManager *services.ClientManager) error {
	w, ok := clientManager.GetClient(userId)
	if !ok {
		log.Println("Client not found")
		return nil
	}

	f, ok := w.(http.Flusher)
	if !ok {
		log.Println("Response writer does not implement http.Flusher")
		return nil
	}

	data, _ := json.Marshal(map[string]string{
		"message": message,
		"userId":  userId,
	})

	_, err := w.Write([]byte("data: " + string(data) + "\n\n"))
	if err != nil {
		log.Println("Error writing to response writer:", err)
		return err
	}
	f.Flush()
	return nil
}

func sendAppToClient(app models.App, clientManager *services.ClientManager) error {
	client, ok := clientManager.GetClient(app.UserID)
	if !ok {
		log.Println("Client not found")
		return nil
	}

	appJson, err := json.Marshal(&app)
	if err != nil {
		log.Println("Error marshalling JSON:", err)
		return err
	}

	f, ok := client.(http.Flusher)
	if !ok {
		log.Println("Client does not implement http.Flusher")
		return err
	}
	_, err = client.Write([]byte("data: " + string(appJson) + "\n\n"))
	if err != nil {
		log.Println("Error sending event:", err)
		return err
	}
	f.Flush()

	return nil
}

func getLiveRequests(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	userId, ok := r.Context().Value(middlewares.UserIDKey).(string)
	if !ok {
		services.AppError("UserID not found in context", 500, w)
		return
	}
	log.Println("User connected:", userId)

	clientManager := services.NewClientManager()
	clientManager.AddClient(userId, w)
	defer clientManager.RemoveClient(userId)

	if err := sendMessage("warmup", userId, clientManager); err != nil {
		return
	}

	appCh := make(chan event.DataEvent)
	// defer close(appCh)

	event.EB.Subscribe("app", appCh)

	type App = models.App

	heartbeatTicker := time.NewTicker(30 * time.Second)
	defer heartbeatTicker.Stop()

	for {
		select {
		case appEvent := <-appCh:
			app, ok := appEvent.Data.(App)
			if !ok {
				log.Println("Interface does not hold type App")
				return
			}
			err := sendAppToClient(app, clientManager)
			if err != nil {
				services.AppError(err.Error(), 500, w)
				return
			}
		case <-heartbeatTicker.C:
			err := sendMessage("heartbeat", userId, clientManager)
			if err != nil {
				log.Println("Error sending heartbeat: ", err)
				return
			}
		case <-r.Context().Done():
			log.Println("Client disconnected")
		}
	}
}

func GetLiveRequestsRoute(router *mux.Router) {
	router.HandleFunc("/get-live", getLiveRequests).Methods("GET")
}
