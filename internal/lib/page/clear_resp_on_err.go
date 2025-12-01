package page

func ClearRespOnErr[T any](err error, resp **T) {
	if err != nil {
		*resp = nil
	}
}