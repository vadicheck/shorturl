package shorten

type Request struct {
	URL string `json:"url"`
}

type Response struct {
	Result string `json:"result"`
}

type ResponseError struct {
	Error string `json:"error"`
}

func NewError(err string) ResponseError {
	return ResponseError{
		Error: err,
	}
}
