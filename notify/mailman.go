package notify

type Mailman interface {
	Enqueue(to string, template Template, args ...interface{}) error
}
