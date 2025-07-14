package metric

import "testing"

func BenchmarkCounter_Update(b *testing.B) {
	c := MustNewCounter("counter1", 0)

	for i := 0; i < b.N; i++ {
		c.Update(1)
	}
}
