package xmlutil

import (
	"encoding/xml"
	"time"
)

func (t Time) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	e.EncodeElement(time.Time(t).Format(FormatISO8601), start)

	return nil
}

func (t *Time) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var str string

	d.DecodeElement(&str, &start)
	parsed, err := time.Parse(FormatISO8601, str)
	if err != nil {
		return err
	}

	*t = Time(parsed)

	return nil
}
