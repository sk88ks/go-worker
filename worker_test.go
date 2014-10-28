package worker

import (
	"errors"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestWrap(t *testing.T) {
	Convey("Given a not function", t, func() {
		f := "test"

		Convey("When args num is same as function args len", func() {
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
				So(res[0].(string), ShouldEqual, "test2")

			})
		})

		Convey("When args num is same as function args len", func() {
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

func TestNewManager(t *testing.T) {
	Convey("Given workerNum", t, func() {
		wn := 5

		Convey("When create new manager", func() {
			m := NewManager(wn)

			Convey("Then returned is Manager", func() {
				So(m, ShouldHaveSameTypeAs, &Manager{})

			})
		})
	})
}

func TestAdd(t *testing.T) {
	Convey("Given function and Manager", t, func() {
		f := "test"

		m := NewManager(5)

		Convey("When Add the function", func() {
			m.Add("test_1", f, "test", 0, []string{"0", "1", "2"})

			Convey("Then", func() {
				ne := m.GetNotExecute()
				So(ne[0], ShouldEqual, "test_1")
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

		m := NewManager(5)

		Convey("When Add the function", func() {
			m.Add("test_1", f, "test", 0, []string{"0", "1", "2"})

			Convey("Then", func() {
			})
		})
	})
}

func TestStart(t *testing.T) {
	Convey("Given function and Manager", t, func() {
		f := func(a string, b int, c []string) (string, error) {
			index := b
			if len(c) <= b {
				return "", errors.New("test")
			}
			return a + c[index], nil
		}

		m := NewManager(5)

		Convey("When Add the function", func() {
			m.Add("test_1", f, "test", 0, []string{"0", "1", "2"})
			m.Start()

			Convey("Then", func() {
			})
		})
	})
}

func TestAddFail(t *testing.T) {
	Convey("Given function and Manager", t, func() {
		f := func(a string, b int, c []string) (string, error) {
			index := b
			if len(c) <= b {
				return "", errors.New("test")
			}
			return a + c[index], nil
		}

		m := NewManager(5)

		Convey("When Add the function", func() {
			m.Add("test_1", f, "test", 10, []string{"0", "1", "2"})
			errMsg := make(chan string)
			m.AddFail(func(p *Process) {
				errMsg <- p.Error.Error()
			})
			m.Start()

			Convey("Then err is returned", func() {
				msg := <-errMsg
				So(msg, ShouldEqual, "test")
			})
		})
	})
}

func TestAddSuccess(t *testing.T) {
	Convey("Given function and Manager", t, func() {
		f := func(a string, b int, c []string) (string, error) {
			index := b
			if len(c) <= b {
				return "", errors.New("test")
			}
			return a + c[index], nil
		}

		m := NewManager(5)

		Convey("When Add the function", func() {
			m.Add("test_1", f, "test", 0, []string{"0", "1", "2"})
			successMsg := make(chan string)
			m.AddSuccess(func(p *Process) {
				successMsg <- p.Result[0].(string)
			})
			m.Start()

			Convey("Then err is returned", func() {
				msg := <-successMsg
				So(msg, ShouldEqual, "test0")
			})
		})
	})
}

func TestEnd(t *testing.T) {
	Convey("Given function and Manager", t, func() {
		f := func(a string, b int, c []string) (string, error) {
			index := b
			if len(c) <= b {
				return "", errors.New("test")
			}
			return a + c[index], nil
		}

		m := NewManager(5)
		m.Add("test_1", f, "test", 0, []string{"0", "1", "2"})
		m.Start()

		Convey("When retrieves results", func() {
			res := m.End()

			Convey("Then err is returned", func() {
				So(res[0].Result[0].(string), ShouldEqual, "test0")
			})
		})
	})
}

func TestEndWithFailStop(t *testing.T) {
	Convey("Given function and Manager", t, func() {
		f := func(a string, b int, c []string) (string, error) {
			index := b
			if len(c) <= b {
				return "", errors.New("test")
			}
			return a + c[index], nil
		}

		m := NewManager(5)
		m.Add("test_1", f, "test", 0, []string{"0", "1", "2"})

		Convey("When retrieves results", func() {
			m.Start()
			res, err := m.EndWithFailStop()

			Convey("Then err is returned", func() {
				So(err, ShouldBeNil)
				So(res[0].Result[0].(string), ShouldEqual, "test0")
			})
		})

		Convey("When return err", func() {
			m.Add("test_1", f, "test", 10, []string{"0", "1", "2"})
			m.Start()
			_, err := m.EndWithFailStop()

			Convey("Then err is returned", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "test")
			})
		})
	})
}