package usermodel

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Role string

const (
	UserRole  Role = "user"
	AdminRole Role = "admin"
)

type User struct {
	ID            uuid.UUID  `gorm:"primaryKey;type:uuid;" json:"id"`
	FullName      string     `gorm:"column:full_name;type:varchar(100);not null" json:"fullName"`
	Email         string     `gorm:"column:email;type:varchar(50);unique;not null" json:"email"`
	Password      string     `gorm:"column:password;type:varchar(255);" json:"password"`
	PasswordHash  string     `gorm:"column:password_hash;type:varchar(255);not null" json:"-"`
	Salt          string     `gorm:"column:salt;type:varchar(255);" json:"-"`
	Role          Role       `gorm:"column:role;type:varchar(10);default:user" json:"role"`
	EmailVerified bool       `gorm:"column:email_verified;default:false" json:"emailVerified"`
	LastLogin     *time.Time `gorm:"column:last_login" json:"lastLogin,omitempty"`
	CreatedAt     time.Time  `gorm:"column:created_at;default:now()" json:"createdAt"`
	UpdatedAt     time.Time  `gorm:"column:updated_at;default:now()" json:"updatedAt"`
}

func (User) TableName() string { return "user" }

// BeforeCreate hook is retained for UUID generation
func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	u.ID = uuid.New()
	return
}

// PublicUser is a struct used for sending user data to the client.
type PublicUser struct {
	FullName      string    `json:"fullName"`
	Email         string    `json:"email"`
	EmailVerified bool      `json:"emailVerified"`
	Role          string    `json:"role"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

// ToPublicUser converts a User instance to a PublicUser instance.
func (u *User) ToPublicUser() PublicUser {
	return PublicUser{
		FullName:      u.FullName,
		Email:         u.Email,
		EmailVerified: u.EmailVerified,
		Role:          string(u.Role),
		CreatedAt:     u.CreatedAt,
		UpdatedAt:     u.UpdatedAt,
	}
}

type UserUpdatePayload struct {
	FullName string `json:"fullName"`
	Email    string `json:"email"`
}
