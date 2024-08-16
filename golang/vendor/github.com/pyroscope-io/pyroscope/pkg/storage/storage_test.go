package storage

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pyroscope-io/pyroscope/pkg/config"
	"github.com/pyroscope-io/pyroscope/pkg/storage/tree"
	"github.com/pyroscope-io/pyroscope/pkg/testing"
)

// 21:22:08      air |  (time.Duration) 10s,
// 21:22:08      air |  (time.Duration) 1m40s,
// 21:22:08      air |  (time.Duration) 16m40s,
// 21:22:08      air |  (time.Duration) 2h46m40s,
// 21:22:08      air |  (time.Duration) 27h46m40s,
// 21:22:08      air |  (time.Duration) 277h46m40s,
// 21:22:08      air |  (time.Duration) 2777h46m40s,
// 21:22:08      air |  (time.Duration) 27777h46m40s

var (
	s  *Storage
	s2 *Storage
)

var _ = Describe("storage package", func() {
	testing.WithConfig(func(cfg **config.Config) {
		JustBeforeEach(func() {
			var err error
			s, err = New(*cfg)
			Expect(err).ToNot(HaveOccurred())
		})

		Context("smoke tests", func() {
			Context("simple 10 second write", func() {
				It("works correctly", func() {
					tree := tree.New()
					tree.Insert([]byte("a;b"), uint64(1))
					tree.Insert([]byte("a;c"), uint64(2))
					st := testing.SimpleTime(10)
					et := testing.SimpleTime(19)
					st2 := testing.SimpleTime(0)
					et2 := testing.SimpleTime(30)
					key, _ := ParseKey("foo")

					err := s.Put(st, et, key, tree, "testspy", 100)
					Expect(err).ToNot(HaveOccurred())

					t2, _, _, _, err := s.Get(st2, et2, key)

					Expect(err).ToNot(HaveOccurred())
					Expect(t2).ToNot(BeNil())
					Expect(t2.String()).To(Equal(tree.String()))
					Expect(s.Close()).ToNot(HaveOccurred())
				})
			})
			Context("simple 20 second write", func() {
				It("works correctly", func() {
					tree := tree.New()
					tree.Insert([]byte("a;b"), uint64(2))
					tree.Insert([]byte("a;c"), uint64(4))
					st := testing.SimpleTime(10)
					et := testing.SimpleTime(29)
					st2 := testing.SimpleTime(0)
					et2 := testing.SimpleTime(30)
					key, _ := ParseKey("foo")

					err := s.Put(st, et, key, tree, "testspy", 100)
					Expect(err).ToNot(HaveOccurred())

					t2, _, _, _, err := s.Get(st2, et2, key)

					Expect(err).ToNot(HaveOccurred())
					Expect(t2).ToNot(BeNil())
					Expect(t2.String()).To(Equal(tree.String()))
					Expect(s.Close()).ToNot(HaveOccurred())
				})
			})

			It("persist data between restarts", func() {
				tree := tree.New()
				tree.Insert([]byte("a;b"), uint64(1))
				tree.Insert([]byte("a;c"), uint64(2))
				st := testing.SimpleTime(10)
				et := testing.SimpleTime(19)
				st2 := testing.SimpleTime(0)
				et2 := testing.SimpleTime(30)
				key, _ := ParseKey("foo")

				err := s.Put(st, et, key, tree, "testspy", 100)
				Expect(err).ToNot(HaveOccurred())

				t2, _, _, _, err := s.Get(st2, et2, key)

				Expect(err).ToNot(HaveOccurred())
				Expect(t2).ToNot(BeNil())
				Expect(t2.String()).To(Equal(tree.String()))
				Expect(s.Close()).ToNot(HaveOccurred())

				s2, err = New(*cfg)
				Expect(err).ToNot(HaveOccurred())

				t3, _, _, _, err := s2.Get(st2, et2, key)
				Expect(err).ToNot(HaveOccurred())
				Expect(t3).ToNot(BeNil())
				Expect(t3.String()).To(Equal(tree.String()))
			})
		})
	})
})