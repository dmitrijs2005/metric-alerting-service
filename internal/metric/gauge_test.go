package metric

import "testing"

func BenchmarkGauge_Update(b *testing.B) {
	c := MustNewGauge("gauge1", 0)

	for i := 0; i < b.N; i++ {
		c.Update(i)
	}
}
