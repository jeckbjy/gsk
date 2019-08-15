package sync

type Cron interface {
	Schedule() error
}
