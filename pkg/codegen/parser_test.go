package codegen

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFormatCode(t *testing.T) {
	src := `
package main
import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"github.com/doordash/oapi-codegen/v3/pkg/codegen"
	"github.com/doordash/oapi-codegen/v3/pkg/codegen/ast"
)
func main() {
	fmt.Println("Hello, World!")
}
`

	expected := `package main

import (
	"fmt"
)

func main() {
	fmt.Println("Hello, World!")
}
`
	res, err := FormatCode(src)
	require.NoError(t, err)
	require.Equal(t, expected, res)
}

func TestOptimizeImports(t *testing.T) {
	src := `
package main
import (
	"fmt"
	"foo"
	"bar"
)
func main() {
	fmt.Println("Hello, World!")
}
`

	expected := `package main

import (
	"fmt"
)

func main() {
	fmt.Println("Hello, World!")
}
`
	res, err := optimizeImports([]byte(src))
	require.NoError(t, err)
	require.Equal(t, expected, string(res))
}
