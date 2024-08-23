package request

import (
	"log"
	"strconv"
	"time"

	"github.com/Tibz-Dankan/keep-active/internal/events"
	"github.com/Tibz-Dankan/keep-active/internal/models"

	"github.com/Tibz-Dankan/keep-active/internal/services"
)

// Makes request for the app
func MakeAppRequest(app models.App) {
	ok, err := validateApp(app)
	if err != nil {
		log.Println("Error validating the app: ", err)
	}
	if !ok {
		log.Println("Couldn't make request: ", app.Name)
		return
	}

	appRequestProgress := services.AppRequestProgress{App: app, InProgress: true}
	services.UserAppMem.Add(app.UserID, appRequestProgress)

	events.EB.Publish("appRequestProgress", appRequestProgress)

	response, err := services.MakeHTTPRequest(app.URL)
	if err != nil {
		log.Println("Request error:", err)
		return
	}

	request := models.Request{
		AppID:      app.ID,
		StatusCode: response.StatusCode,
		Duration:   response.RequestTimeMS,
		StartedAt:  response.StartedAt,
	}
	log.Printf("Request statusCode: %d Duration: %d URL: %s\n", request.StatusCode, request.Duration, app.URL)

	request, err = request.Create(request)
	if err != nil {
		log.Println("Error saving request:", err)
	}

	appRequestProgress.Request = []models.Request{request}
	appRequestProgress.InProgress = false
	services.UserAppMem.Add(app.UserID, appRequestProgress)

	events.EB.Publish("appRequestProgress", appRequestProgress)
}

// Validates the app's eligibility for making requests
func validateApp(app models.App) (bool, error) {
	if app.IsDisabled {
		return false, nil
	}

	hasLastRequest := len(app.Request) > 0

	// Check and validate requestInterval
	if len(app.RequestTime) == 0 {
		if !hasLastRequest {
			return true, nil
		}
		log.Println("App doesn't have requestTime")
		currentTime := time.Now()
		location := currentTime.Location().String()
		appDate := services.Date{TimeZone: location, ISOStringDate: app.Request[0].StartedAt.String()}

		currentAppTime, _ := appDate.CurrentTime()
		lastRequestStartedAt, _ := appDate.ISOTime()
		timeDiff := currentAppTime.Sub(lastRequestStartedAt).Minutes()
		requestInterval, err := strconv.Atoi(app.RequestInterval)
		if err != nil {
			log.Println("Error converting string to integer:", err)
		}

		if int(timeDiff) >= requestInterval {
			return true, nil
		}
		return false, nil
	}

	for _, rt := range app.RequestTime {
		// Check and validate requestTime slot

		lastReqStartedAtStr := time.Now().String()
		if hasLastRequest {
			lastReqStartedAtStr = app.Request[0].StartedAt.String()
		}

		appDateStart := services.Date{TimeZone: rt.TimeZone, ISOStringDate: lastReqStartedAtStr, HourMinSec: rt.Start}
		appDateEnd := services.Date{TimeZone: rt.TimeZone, ISOStringDate: lastReqStartedAtStr, HourMinSec: rt.End}

		startTime, _ := appDateStart.HourMinSecTime()
		endTime, _ := appDateEnd.HourMinSecTime()
		currentTimeStart, _ := appDateStart.CurrentTime()
		currentTimeEnd, _ := appDateEnd.CurrentTime()

		isEqualToStartTime := currentTimeStart.Equal(startTime)
		isEqualToEndTime := currentTimeEnd.Equal(endTime)
		isGreaterThanStartTime := currentTimeStart.After(startTime)
		isLessThanEndTime := currentTimeEnd.Before(endTime)

		isWithinRequestTimeRange := isGreaterThanStartTime && isLessThanEndTime

		if isEqualToStartTime || isEqualToEndTime || isWithinRequestTimeRange {
			// Check and validate requestInterval
			log.Println("App time frame is correct")
			if !hasLastRequest {
				return true, nil
			}
			lastRequestStartedAt, _ := appDateStart.ISOTime()
			timeDiff := currentTimeStart.Sub(lastRequestStartedAt).Minutes()
			requestInterval, err := strconv.Atoi(app.RequestInterval)
			if err != nil {
				log.Println("Error converting string to integer:::", err)
			}

			if int(timeDiff) >= requestInterval {
				return true, nil
			}
		}
	}

	return false, nil
}
