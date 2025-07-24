package tm1go

type ValueArray[T any] struct {
	Count int `json:"@odata.count"`
	Value []T `json:"value"`
}

type Value[T any] struct {
	Value T `json:"value"`
}
