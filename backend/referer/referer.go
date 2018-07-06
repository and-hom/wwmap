package referer

type RefererStorage interface {
	PutHttp(url string)
	PutHttps(url string)
	List() []string
}

type Dummy struct {

}

func (this Dummy) PutHttp(url string) {}

func (this Dummy) PutHttps(url string) {}

func (this Dummy) List() []string {
	return []string{}
}