package core

import "time"

type OpenKitObject interface {
	closeAt(timestamp time.Time)
	close()
}
