package testutil

import (
	"time"

	"github.com/google/uuid"
	"hpt.space/bltfs/util/xmlutil"
)

var TestUUID = mustParseUUID("df925be0-44c0-4e49-af6a-7c3aa5b36d35")

func mustParseUUID(s string) uuid.UUID {
	parsed, err := uuid.Parse(s)
	if err != nil {
		panic(err)
	}

	return parsed
}

var TestTime = mustParse(xmlutil.FormatISO8601, "2017-03-07T13:27:38.192689471Z")

func mustParse(layout, value string) xmlutil.Time {
	parsed, err := time.Parse(layout, value)
	if err != nil {
		panic(err)
	}

	return xmlutil.Time(parsed)
}
