package flow

type Flow interface {
	GetType() Type
	GetRequestURL() string
}
