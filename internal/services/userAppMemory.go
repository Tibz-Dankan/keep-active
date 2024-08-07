package services

import (
	"sync"

	"github.com/Tibz-Dankan/keep-active/internal/models"
)

type AppRequestProgress struct {
	models.App
	InProgress bool `json:"inProgress"`
}

type UserAppMemory struct {
	user map[string][]AppRequestProgress
	sync.RWMutex
}

func (uam *UserAppMemory) Add(userId string, app AppRequestProgress) {
	uam.Lock()
	defer uam.Unlock()

	if prevApps, found := uam.user[userId]; found {
		replaced := false
		for i, existingApp := range prevApps {
			if existingApp.ID == app.ID {
				prevApps[i] = app
				replaced = true
				uam.user[userId] = prevApps
				break
			}
		}
		if !replaced {
			uam.user[userId] = append(prevApps, app)
		}
		return
	}

	uam.user[userId] = []AppRequestProgress{app}
}

func (uam *UserAppMemory) Delete(userId string) {
	uam.Lock()
	defer uam.Unlock()
	delete(uam.user, userId)
}

func (uam *UserAppMemory) DeleteAll() {
	uam.Lock()
	defer uam.Unlock()
	for userId := range uam.user {
		delete(uam.user, userId)
	}
}

func (uam *UserAppMemory) Get(userId string) ([]AppRequestProgress, bool) {
	uam.Lock()
	defer uam.Unlock()
	apps, ok := uam.user[userId]

	return apps, ok
}

var UserAppMem = &UserAppMemory{
	user: make(map[string][]AppRequestProgress),
}
