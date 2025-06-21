package str

import (
	"fmt"
	"testing"
)

func TestCompareVersions(t *testing.T) {
	data := CompareVersions("v1.1.03", "v1.1.03")
	fmt.Println(data)
}
