// notemodel defines the data structures used in the application.
package notemodel

import (
	"time"

	"github.com/google/uuid"
)

// Note represents a note created by a software engineer to solve a problem.
type Note struct {
	ID        uuid.UUID `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	UserID    uuid.UUID `gorm:"type:uuid"`
	Title     string    `gorm:"type:varchar(255);not null"`
	Problem   string    `gorm:"type:text;not null"`
	Approach  string    `gorm:"type:text;not null"`
	Solution  string    `gorm:"type:text;not null"`
	ExtraNote string    `gorm:"type:text"`
	CreatedAt time.Time `gorm:"default:now()"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}
