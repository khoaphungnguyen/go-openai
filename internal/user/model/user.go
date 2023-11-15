package modeluser

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
	FirstName     string     `gorm:"type:varchar(50);not null" json:"first_name"`
	LastName      string     `gorm:"type:varchar(50);not null" json:"last_name"`
	Email         string     `gorm:"type:varchar(50);unique;not null" json:"email"`
	Password      string     `gorm:"type:varchar(255);not null" json:"-"`
	Salt          string     `gorm:"type:varchar(255);" json:"-"`
	Role          Role       `gorm:"type:varchar(10);default:user" json:"role"`
	EmailVerified bool       `gorm:"default:false" json:"email_verified"`
	LastLogin     *time.Time `json:"last_login"`
	CreatedAt     time.Time  `gorm:"default:now()" json:"created_at"`
	UpdatedAt     time.Time  `gorm:"default:now()" json:"updated_at"`
}

// BeforeCreate hook is retained for UUID generation
func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	u.ID = uuid.New()
	return
}
