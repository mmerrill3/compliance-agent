package tasks

type Task interface {
	Build(cmd string) (ouptut string, err error)
	GetFileName() string
}
