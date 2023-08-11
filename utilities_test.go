package main

import (
	"net/http"
	"testing"
	"time"
)

func TestTimeCheck(t *testing.T) {
	localDateString := "02/01/2006"

	//Set up float times from easy to read dates
	earlyStartDate, _ := time.Parse(localDateString, "17/10/1998")
	lateStartDate, _ := time.Parse(localDateString, "01/01/2099")
	earlyStartTime := float64(earlyStartDate.Unix())
	lateStartTime := float64(lateStartDate.Unix())
	nowTime := float64(time.Now().Unix())
	var nilTime time.Time

	// Set min and max boundaries (yesterday and tomorrow)
	yesterday := time.Now().AddDate(0, 0, -1)
	tomorrow := time.Now().AddDate(0, 0, 1)

	// Check for no replacement
	min, max := timeCheck(nowTime, yesterday, tomorrow)
	if min == yesterday && max == tomorrow {
		t.Log("No Change Made - Correct")
	} else {
		t.Errorf("Unecessarily Changed Min/Max Times\n\tMin: %v\n\tMax: %v", min == yesterday, max == tomorrow)
	}

	// Check for 0 replacement (where there is no min time)
	min, _ = timeCheck(earlyStartTime, nilTime, nilTime)
	if min.Compare(earlyStartDate) == 0 {
		t.Log("Min Time Initialised - Correct")
	} else {
		t.Error("Min Time Not Initialised")
	}

	// Check for min replacement
	min, _ = timeCheck(earlyStartTime, yesterday, tomorrow)
	if min.Compare(earlyStartDate) == 0 {
		t.Log("Changed Min Time - Correct")
	} else {
		t.Log(min)
		t.Log(earlyStartDate)
		t.Error("Did not change Min Time")
	}

	// // Check for max replacement
	_, max = timeCheck(lateStartTime, yesterday, tomorrow)
	if max.Compare(lateStartDate) == 0 {
		t.Log("Changed Max Time - Correct")
	} else {
		t.Log(max)
		t.Log(lateStartDate)
		t.Error("Did not change Max Time")
	}
}

func TestGetSSEResponse(t *testing.T) {
	resp, _, _ := getSSEResponse()
	if resp.Request.URL.String() == "https://stream.upfluence.co/stream" {
		t.Log("Correct URL Requested")
	} else {
		t.Log(resp.Request.URL.String())
		t.Error("Incorrect URL Requested")
	}

	if resp.StatusCode == http.StatusOK {
		t.Log("Request Accepted")
	} else {
		t.Log(resp.StatusCode)
		t.Error("Request Refused")
	}
}

func TestExtractNumericKey(t *testing.T) {
	// testing data with a 'likes' key but no 'retweets' key, and some extra keys
	// All the values are float64 - this is what they are read as default, so had to be forced here
	data := map[string]any{
		"id":        100536070,
		"title":     "A Test Datastream",
		"likes":     float64(15),
		"comments":  float64(0),
		"timestamp": float64(1691726344),
	}

	val, _ := extractNumericKey(data, "likes")
	if val == 15 {
		t.Log("Correct Likes Value Extracted")
	} else {
		t.Logf("%v: %T", val, val)
		t.Error("Incorrect Likes Value Extracted")
	}

	// Check for 0 extraction
	data["likes"] = 0
	val, _ = extractNumericKey(data, "likes")
	if val == 0 {
		t.Log("Correct 0 Value Extracted")
	} else {
		t.Error("Incorrect 0 Value Extracted")
	}

	// Check for non-existant key extraction
	val, _ = extractNumericKey(data, "likes")
	if val == 0 {
		t.Log("Correct non-existant key extracted")
	} else {
		t.Error("Incorrect Non-Existant Key Value Extracted")
	}
}
