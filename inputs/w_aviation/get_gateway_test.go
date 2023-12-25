package w_aviation_test

import (
	"fmt"
	"testing"

	"flashcat.cloud/categraf/inputs/w_aviation"
)

func TestGetGateway(t *testing.T) {
	gw := w_aviation.GetGateway()
	fmt.Println(gw)
}
