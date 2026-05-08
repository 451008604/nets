package benchmark

import (
	"fmt"
	"github.com/451008604/nets"
	"testing"
)

func BenchmarkPack(b *testing.B) {
	sizes := []int{16, 64, 256, 1024, 4096}
	dp := nets.NewDataPack()
	for _, size := range sizes {
		data := make([]byte, size)
		for i := range data {
			data[i] = byte(i % 256)
		}
		msg := nets.NewMsgPackage(1001, data)
		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				packed := dp.Pack(msg)
				_ = packed
			}
		})
	}
}

func BenchmarkUnPack(b *testing.B) {
	sizes := []int{16, 64, 256, 1024, 4096}
	dp := nets.NewDataPack()
	for _, size := range sizes {
		data := make([]byte, size)
		for i := range data {
			data[i] = byte(i % 256)
		}
		msg := nets.NewMsgPackage(1001, data)
		packed := dp.Pack(msg)
		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				unpacked := dp.UnPack(packed)
				nets.PutMessage(unpacked)
			}
		})
	}
}

func BenchmarkPackUnPackRoundTrip(b *testing.B) {
	sizes := []int{16, 64, 256, 1024, 4096}
	dp := nets.NewDataPack()
	for _, size := range sizes {
		data := make([]byte, size)
		for i := range data {
			data[i] = byte(i % 256)
		}
		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				msg := nets.NewMsgPackage(1001, data)
				packed := dp.Pack(msg)
				unpacked := dp.UnPack(packed)
				nets.PutMessage(unpacked)
			}
		})
	}
}
