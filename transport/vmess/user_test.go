package vmess

import (
	"fmt"
	"testing"

	"github.com/gofrs/uuid"
)

func TestNewID(t *testing.T) {
	uid, err := uuid.FromString("803277e5-64ef-48f9-8fd5-de82481ee789")
	if err != nil {
		t.Error(err)
	}
	newid := newID(&uid)
	fmt.Println(newid.CmdKey)
	ids := newAlterIDs(newid, 1)
	for _, id := range ids {
		fmt.Println(id.UUID.String())
	}
}
