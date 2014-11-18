package worker

import (
	"errors"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestWrap(t *testing.T) {
	Convey("Given a not function", t, func() {
		f := "test"

		Convey("When args num is not same as function args len", func() {
			wf := wrap(f, "test", 1, []string{"1", "2", "3"})

			Convey("Then returned ProcessFunc", func() {
				So(wf, ShouldBeNil)

			})
		})
	})

	Convey("Given a function", t, func() {
		f := func(a string, b int, c []string) (string, error) {
			index := b
			if len(c) <= b {
				return "", errors.New("test")
			}
			return a + c[index], nil
		}

		Convey("When args num is same as function args len", func() {
			wf := wrap(f, "test", 1, []string{"1", "2", "3"})

			Convey("Then returned ProcessFunc", func() {
				So(wf, ShouldNotBeNil)
				res, err := wf()
				So(err, ShouldBeNil)
				So(res.(string), ShouldEqual, "test2")

			})
		})

		Convey("When args num is same as function args len for returning err", func() {
			wf := wrap(f, "test", 10, []string{"1", "2", "3"})

			Convey("Then returned error", func() {
				So(wf, ShouldNotBeNil)
				_, err := wf()
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "test")

			})
		})

		Convey("When args num is  different with  function args len", func() {
			wf := wrap(f, "test", 1, []string{"1", "2", "3"}, "invalid_arg")

			Convey("Then returned ProcessFunc", func() {
				So(wf, ShouldBeNil)

			})
		})
	})
}

func TestNew(t *testing.T) {
	Convey("Given workerNum", t, func() {
		wn := 5

		Convey("When create new manager", func() {
			m := New(wn)

			Convey("Then returned is Manager", func() {
				So(m, ShouldHaveSameTypeAs, &Manager{})

			})
		})
	})
}

func TestAdd(t *testing.T) {
	Convey("Given function and Manager", t, func() {
		f := "test"

		m := New(5)

		Convey("When Add the function", func() {
			m.Add("test_add", f, "test", 0, []string{"0", "1", "2"})

			Convey("Then", func() {
				ne := m.GetNotExecute()
				So(ne[0], ShouldEqual, "test_add")
			})
		})
	})

	Convey("Given function and Manager", t, func() {
		f := func(a string, b int, c []string) (string, error) {
			index := b
			if len(c) <= b {
				return "", errors.New("test")
			}
			return a + c[index], nil
		}

		m := New(5)

		Convey("When Add the function", func() {
			m.Add("test_add", f, "test", 0, []string{"0", "1", "2"})

			Convey("Then", func() {
			})
		})
	})
}

func TestRun(t *testing.T) {
	Convey("Given function and Manager", t, func() {
		f := func(a string, b int, c []string) (string, error) {
			index := b
			if len(c) <= b {
				return "", errors.New("test")
			}
			return a + c[index], nil
		}

		m := New(5)
		m.Add("TestRun1", f, "test1", 0, []string{"0", "1", "2"})
		m.Add("TestRun2", f, "test2", 0, []string{"0", "1", "2"})

		Convey("When retrieves results", func() {
			type Result struct {
				TestRun1 string
				TestRun2 string
			}
			res := &Result{}
			m.Run(res)

			Convey("Then err is returned", func() {
				So(res.TestRun1, ShouldEqual, "test10")
				So(res.TestRun2, ShouldEqual, "test20")
			})
		})
	})
}

func TestFail(t *testing.T) {
	Convey("Given function and Manager", t, func() {
		f := func(a string, b int, c []string) (string, error) {
			index := b
			if len(c) <= b {
				return "", errors.New("test")
			}
			return a + c[index], nil
		}

		m := New(5)

		Convey("When Add the function", func() {
			m.Add("test_fail", f, "test", 10, []string{"0", "1", "2"})
			var errMsg string
			m.Fail(func(p *Process) {
				errMsg = p.Error.Error()
			})
			m.Run(nil)

			Convey("Then err is returned", func() {
				So(errMsg, ShouldEqual, "test")
			})
		})
	})
}

func TestSuccess(t *testing.T) {
	Convey("Given function and Manager", t, func() {
		f := func(a string, b int, c []string) (string, error) {
			index := b
			if len(c) <= b {
				return "", errors.New("test")
			}
			return a + c[index], nil
		}

		m := New(5)

		Convey("When Add the function", func() {
			m.Add("test_1", f, "test", 0, []string{"0", "1", "2"})
			var successMsg string
			m.Success(func(p *Process) {
				successMsg = p.Result.(string)
			})
			m.Run(nil)

			Convey("Then err is returned", func() {
				So(successMsg, ShouldEqual, "test0")
			})
		})
	})
}
