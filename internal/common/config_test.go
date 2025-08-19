package common

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_jsonConfigEnv(t *testing.T) {
	t.Run("returns value when CONFIG is set", func(t *testing.T) {
		t.Setenv("CONFIG", "/etc/conf.json")
		assert.Equal(t, "/etc/conf.json", JsonConfigEnv())
	})

	t.Run("returns empty when CONFIG is unset", func(t *testing.T) {
		t.Setenv("CONFIG", "")
		assert.Empty(t, JsonConfigEnv())
	})
}

func Test_jsonConfigFlags(t *testing.T) {
	origArgs := os.Args
	t.Cleanup(func() { os.Args = origArgs })

	t.Run("short -c with value", func(t *testing.T) {
		os.Args = []string{"testbin", "-c", "/path/short.json"}
		assert.Equal(t, "/path/short.json", JsonConfigFlags())
	})

	t.Run("long -config with value", func(t *testing.T) {
		os.Args = []string{"testbin", "-config", "/path/long.json"}
		assert.Equal(t, "/path/long.json", JsonConfigFlags())
	})

	t.Run("unknown flags are ignored", func(t *testing.T) {
		os.Args = []string{"testbin", "-x", "1", "-y", "2"}
		assert.Empty(t, JsonConfigFlags())
	})

	t.Run("multiple flags, last wins", func(t *testing.T) {
		os.Args = []string{"testbin", "-c", "/path/1.json", "-config", "/path/2.json"}
		assert.Equal(t, "/path/2.json", JsonConfigFlags())
	})
}
