package errors

type AeonError string

func (e AeonError) Error() string {
	return string(e)
}

const (
	ErrUnitOfWorkRunning        AeonError = "a unit of work is already running"
	ErrTimeInFuture             AeonError = "time cannot be in the future"
	ErrTimeWithinCompletedUnit  AeonError = "provided time is within a previously completed unit of work"
	ErrNoUnitOfWorkRunning      AeonError = "no unit of work is running"
	ErrStopTimeBeforeStartTime  AeonError = "stop time cannot be before the start time"
	ErrCompensationOnNonWorkDay AeonError = "compensation on a non-work day is not allowed"
)
