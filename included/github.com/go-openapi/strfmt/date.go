// Copyright 2015 go-swagger maintainers
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package strfmt

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"time"

	"github.com/aaron-prindle/krmapiserver/included/github.com/globalsign/mgo/bson"
	"github.com/aaron-prindle/krmapiserver/included/github.com/mailru/easyjson/jlexer"
	"github.com/aaron-prindle/krmapiserver/included/github.com/mailru/easyjson/jwriter"
)

func init() {
	d := Date{}
	// register this format in the default registry
	Default.Add("date", &d, IsDate)
}

// IsDate returns true when the string is a valid date
func IsDate(str string) bool {
	_, err := time.Parse(RFC3339FullDate, str)
	return err == nil
}

const (
	// RFC3339FullDate represents a full-date as specified by RFC3339
	// See: http://goo.gl/xXOvVd
	RFC3339FullDate = "2006-01-02"
)

// Date represents a date from the API
//
// swagger:strfmt date
type Date time.Time

// String converts this date into a string
func (d Date) String() string {
	return time.Time(d).Format(RFC3339FullDate)
}

// UnmarshalText parses a text representation into a date type
func (d *Date) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		return nil
	}
	dd, err := time.Parse(RFC3339FullDate, string(text))
	if err != nil {
		return err
	}
	*d = Date(dd)
	return nil
}

// MarshalText serializes this date type to string
func (d Date) MarshalText() ([]byte, error) {
	return []byte(d.String()), nil
}

// Scan scans a Date value from database driver type.
func (d *Date) Scan(raw interface{}) error {
	switch v := raw.(type) {
	case []byte:
		return d.UnmarshalText(v)
	case string:
		return d.UnmarshalText([]byte(v))
	case time.Time:
		*d = Date(v)
		return nil
	case nil:
		*d = Date{}
		return nil
	default:
		return fmt.Errorf("cannot sql.Scan() strfmt.Date from: %#v", v)
	}
}

// Value converts Date to a primitive value ready to written to a database.
func (d Date) Value() (driver.Value, error) {
	return driver.Value(d.String()), nil
}

// MarshalJSON returns the Date as JSON
func (d Date) MarshalJSON() ([]byte, error) {
	var w jwriter.Writer
	d.MarshalEasyJSON(&w)
	return w.BuildBytes()
}

// MarshalEasyJSON writes the Date to a easyjson.Writer
func (d Date) MarshalEasyJSON(w *jwriter.Writer) {
	w.String(time.Time(d).Format(RFC3339FullDate))
}

// UnmarshalJSON sets the Date from JSON
func (d *Date) UnmarshalJSON(data []byte) error {
	if string(data) == jsonNull {
		return nil
	}
	l := jlexer.Lexer{Data: data}
	d.UnmarshalEasyJSON(&l)
	return l.Error()
}

// UnmarshalEasyJSON sets the Date from a easyjson.Lexer
func (d *Date) UnmarshalEasyJSON(in *jlexer.Lexer) {
	if data := in.String(); in.Ok() {
		tt, err := time.Parse(RFC3339FullDate, data)
		if err != nil {
			in.AddError(err)
			return
		}
		*d = Date(tt)
	}
}

// GetBSON returns the Date as a bson.M{} map.
func (d *Date) GetBSON() (interface{}, error) {
	return bson.M{"data": d.String()}, nil
}

// SetBSON sets the Date from raw bson data
func (d *Date) SetBSON(raw bson.Raw) error {
	var m bson.M
	if err := raw.Unmarshal(&m); err != nil {
		return err
	}

	if data, ok := m["data"].(string); ok {
		rd, err := time.Parse(RFC3339FullDate, data)
		*d = Date(rd)
		return err
	}

	return errors.New("couldn't unmarshal bson raw value as Date")
}
