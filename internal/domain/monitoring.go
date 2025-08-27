package domain

import "time"

type HTTPMetrics interface {
	IncRequest()
	ObserveRequest(start time.Time)
}
