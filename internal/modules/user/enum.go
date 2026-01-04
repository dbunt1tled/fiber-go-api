package user

import (
	"strconv"
	"strings"

	"github.com/bytedance/sonic"
	"github.com/dbunt1tled/fiber-go-api/pkg/db/dbtype"
)

type Status int

const (
	Inactive Status = iota
	Pending
	Active
)

type Role string

type Roles dbtype.TextArray[Role]

const (
	Admin  Role = "admin"
	Person Role = "person"
)

func (s *Status) String() string {
	switch *s {
	case Inactive:
		return "inactive"
	case Active:
		return "active"
	default:
		return "unknown"
	}
}

func (s Status) MarshalJSON() ([]byte, error) {
	return sonic.ConfigFastest.Marshal(int(s))
}

func (s *Status) UnmarshalJSON(data []byte) error {
	var num int
	if err := sonic.ConfigFastest.Unmarshal(data, &num); err != nil {
		return err
	}
	*s = Status(num)

	return nil
}

func (s Status) MarshalText() ([]byte, error) {
	return []byte(strconv.Itoa(int(s))), nil
}

func (r *Roles) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		*r = nil
		return nil
	}

	parts := strings.Split(string(text), ",")
	res := make(Roles, len(parts))
	for i, p := range parts {
		res[i] = Role(p)
	}
	*r = res
	return nil
}
