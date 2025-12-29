package user

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Address struct {
	Street  string `json:"street"`
	City    string `json:"city"`
	Country string `json:"country"`
	Zip     string `json:"zip"`
}

type User struct {
	ID          uuid.UUID  `db:"id"           json:"id"`
	FirstName   string     `db:"first_name"   json:"firstName"`
	SecondName  string     `db:"second_name"  json:"secondName"`
	Email       string     `db:"email"        json:"email"`
	PhoneNumber string     `db:"phone_number" json:"phoneNumber"`
	Status      Status     `db:"status"       json:"status"`
	Password    string     `db:"password"     json:"password"`
	Roles       Roles      `db:"roles"        json:"roles"`
	Address     *Address   `db:"address"      json:"address"`
	ConfirmedAt *time.Time `db:"confirmed_at" json:"confirmedAt"`
	CreatedAt   time.Time  `db:"created_at"   json:"createdAt"`
	UpdatedAt   time.Time  `db:"updated_at"   json:"updatedAt"`
}

func (u *User) TableName() string  { return "users" }
func (u *User) GetID() uuid.UUID   { return u.ID }
func (u *User) SetID(id uuid.UUID) { u.ID = id }
func (u *User) NextID() *User {
	var err error
	u.ID, err = uuid.NewV7()
	if err != nil {
		panic(fmt.Errorf("failed to generate uuid: %w", err))
	}
	return u
}

func (u *User) WithPassword(password string) *User {
	u.Password = password
	return u
}

func (u *User) Sanitize() {
	u.FirstName = strings.TrimSpace(u.FirstName)
	u.SecondName = strings.TrimSpace(u.SecondName)
	u.Email = strings.TrimSpace(u.Email)
	u.PhoneNumber = strings.TrimSpace(u.PhoneNumber)
	u.Password = strings.TrimSpace(u.Password)
}
