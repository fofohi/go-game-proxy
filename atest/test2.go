package main

func test(params *string) string {
	*params = "test"

	return *params
}

type RecordWriter struct {
}

func (rw RecordWriter) WriteHeader(statusCode *int) *int {
	*statusCode = 2
	return statusCode
}
