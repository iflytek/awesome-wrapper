package daemon

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func Test_DisposedList(t *testing.T) {

	// Only pass t into top-level Convey calls
	Convey("初始化 disposedList", t, func() {
		disposedList := &disposedList{}

		Convey("添加一些值", func() {
			disposedList.set(30)
			disposedList.set(21)
			disposedList.set(99)

			Convey("测试getMin方法是否工作", func() {
				minTmp, minOk := disposedList.getMin()
				So(minTmp, ShouldEqual, 21)
				So(minOk, ShouldEqual, true)
				Convey("剩下的值应该只有30、99：", func() {
					Printf("剩下的值为：%+v\n", disposedList.l)
				})
			})
		})
	})
}
