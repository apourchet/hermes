package hermes

type QueryTemplater func(path string, in interface{}) (string, error)

var DefaultQueryTemplate QueryTemplater = IdentityQueryTemplate

func IdentityQueryTemplate(path string, _ interface{}) (string, error) {
	return path, nil
}
