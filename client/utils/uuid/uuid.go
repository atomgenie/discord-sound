package uuid

import uuidg "github.com/google/uuid"

// Gen Generate UUID
func Gen() string {
	return uuidg.NewString()
}
