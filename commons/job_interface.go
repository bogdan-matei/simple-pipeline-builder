package commons

type JobRun interface {
	Run() error
	ParseFlags() error
}
