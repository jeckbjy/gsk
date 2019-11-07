package httpx

var client = New(nil)

func PostForm(url string, req interface{}, result interface{}, opts ...Option) (*Response, error) {
	return client.Post(url, req, result, append(opts, ContentType(TypeForm))...)
}

func Post(url string, req interface{}, result interface{}, opts ...Option) (*Response, error) {
	return client.Post(url, req, result, opts...)
}

func Get(url string, result interface{}, opts ...Option) (*Response, error) {
	return client.Get(url, result, opts...)
}
