package usermodel

import (
	"time"

	"github.com/google/uuid"
)

// Role defines the type for user roles
type Role string

const (
	UserRole  Role = "user"
	AdminRole Role = "admin"
)

// User represents the user entity as stored in the database.
type User struct {
	ID            uuid.UUID  `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	FullName      string     `gorm:"column:full_name;type:varchar(100);not null"`
	Email         string     `gorm:"column:email;type:varchar(50);unique;not null"`
	PasswordHash  string     `gorm:"column:password_hash;type:varchar(255);not null"`
	Salt          string     `gorm:"column:salt;type:varchar(255);not null"`
	Role          Role       `gorm:"column:role;type:varchar(10);default:user"`
	EmailVerified bool       `gorm:"column:email_verified;default:false"`
	LastLogin     *time.Time `gorm:"column:last_login"`
	DeletedAt     *time.Time `gorm:"index"`
	CreatedAt     time.Time  `gorm:"column:created_at;default:now()"`
	UpdatedAt     time.Time  `gorm:"column:updated_at;default:now()"`
}

func (User) TableName() string {
	return "users"
}

// PublicUser represents a safe-to-expose user model for client-side interactions.
type PublicUser struct {
	ID            uuid.UUID `json:"id"`
	FullName      string    `json:"fullName"`
	Email         string    `json:"email"`
	EmailVerified bool      `json:"emailVerified"`
	Role          string    `json:"role"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

// ToPublicUser converts a User to its public representation.
func (u *User) ToPublicUser() PublicUser {
	return PublicUser{
		ID:            u.ID,
		FullName:      u.FullName,
		Email:         u.Email,
		EmailVerified: u.EmailVerified,
		Role:          string(u.Role),
		CreatedAt:     u.CreatedAt,
		UpdatedAt:     u.UpdatedAt,
	}
}

type AllUser struct {
	FullName      string     `json:"fullName"`
	Email         string     `json:"email"`
	EmailVerified bool       `json:"emailVerified"`
	CreatedAt     time.Time  `json:"createdAt"`
	LastLogin     *time.Time `json:"lastLogin"`
}

// ToPublicUser converts a User to its public representation.
func (u *User) ToAllUser() AllUser {
	return AllUser{
		FullName:      u.FullName,
		Email:         u.Email,
		EmailVerified: u.EmailVerified,
		CreatedAt:     u.CreatedAt,
		LastLogin:     u.LastLogin,
	}
}
