package payload

type Loader interface {
	Ok(method string, data any) map[string]any
	Err(method string, error []error) map[string]any
}
