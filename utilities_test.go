package main

import (
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
