package hermes

import "fmt"

type Resolver func(sni, path string) (string, error)

var DefaultResolver Resolver = func(sni, path string) (string, error) {
	return fmt.Sprintf("%s%s", sni, path), nil
}
