package shortener


type URLStore struct {
	linksMap map[string]string
}


func NewURLStore() *URLStore {
	return &URLStore{
		linksMap: make(map[string]string),
	}
}
