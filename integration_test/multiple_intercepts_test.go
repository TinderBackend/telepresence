package integration_test

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/TinderBackend/telepresence/v2/integration_test/itest"
	"github.com/datawire/dlib/dlog"
)

type multipleInterceptsSuite struct {
	itest.Suite
	itest.MultipleServices
	servicePort   []int
	serviceCancel []context.CancelFunc
}

func init() {
	itest.AddMultipleServicesSuite("", "hello", func(h itest.MultipleServices) suite.TestingSuite {
		return &multipleInterceptsSuite{
			Suite:            itest.Suite{Harness: h},
			MultipleServices: h,
			servicePort:      make([]int, h.ServiceCount()),
			serviceCancel:    make([]context.CancelFunc, h.ServiceCount()),
		}
	})
}

func (s *multipleInterceptsSuite) SetupSuite() {
	s.Suite.SetupSuite()
	ctx := s.Context()
	for i := 0; i < s.ServiceCount(); i++ {
		s.servicePort[i], s.serviceCancel[i] = itest.StartLocalHttpEchoServer(ctx, fmt.Sprintf("%s-%d", s.Name(), i))
	}

	wg := sync.WaitGroup{}
	wg.Add(s.ServiceCount())
	for i := 0; i < s.ServiceCount(); i++ {
		go func(i int) {
			defer wg.Done()
			svc := fmt.Sprintf("%s-%d", s.Name(), i)
			stdout := itest.TelepresenceOk(ctx, "intercept", "--namespace", s.AppNamespace(), svc, "--mount", "false", "--port", strconv.Itoa(s.servicePort[i]))
			s.Contains(stdout, "Using Deployment "+svc)
		}(i)
	}
	wg.Wait()
}

func (s *multipleInterceptsSuite) TearDownSuite() {
	ctx := s.Context()
	for i := 0; i < s.ServiceCount(); i++ {
		itest.TelepresenceOk(ctx, "leave", fmt.Sprintf("%s-%d-%s", s.Name(), i, s.AppNamespace()))
	}
	for _, cancel := range s.serviceCancel {
		if cancel != nil {
			cancel()
		}
	}
}

func (s *multipleInterceptsSuite) Test_Intercepts() {
	ctx := s.Context()
	stdout := itest.TelepresenceOk(ctx, "--namespace", s.AppNamespace(), "list", "--intercepts")
	for i := 0; i < s.ServiceCount(); i++ {
		s.Regexp(fmt.Sprintf(`%s-%d\s*: intercepted`, s.Name(), i), stdout)
	}

	wg := sync.WaitGroup{}
	wg.Add(s.ServiceCount())
	for i := 0; i < s.ServiceCount(); i++ {
		go func(i int) {
			defer wg.Done()
			svc := fmt.Sprintf("%s-%d", s.Name(), i)
			expectedOutput := fmt.Sprintf("%s from intercept at /", svc)
			s.Require().Eventually(
				// condition
				func() bool {
					ip, err := net.DefaultResolver.LookupHost(ctx, svc)
					if err != nil {
						dlog.Infof(ctx, "%v", err)
						return false
					}
					if 1 != len(ip) {
						dlog.Infof(ctx, "Lookup for %s returned %v", svc, ip)
						return false
					}

					dlog.Infof(ctx, "trying %q...", "http://"+svc)
					hc := http.Client{Timeout: 2 * time.Second}
					resp, err := hc.Get("http://" + svc)
					if err != nil {
						dlog.Infof(ctx, "%v", err)
						return false
					}
					defer resp.Body.Close()
					dlog.Infof(ctx, "status code: %v", resp.StatusCode)
					body, err := io.ReadAll(resp.Body)
					if err != nil {
						dlog.Infof(ctx, "%v", err)
						return false
					}
					dlog.Infof(ctx, "body: %q", body)
					return string(body) == expectedOutput
				},
				time.Minute,   // waitFor
				3*time.Second, // polling interval
				`body of %q equals %q`, "http://"+svc, expectedOutput,
			)
		}(i)
	}
	wg.Wait()
}

func (s *multipleInterceptsSuite) Test_ReportsPortConflict() {
	_, stderr, err := itest.Telepresence(s.Context(), "intercept", "--namespace", s.AppNamespace(), "--mount", "false", "--port", strconv.Itoa(s.servicePort[1]), "dummy-name")
	s.Error(err)
	s.Contains(stderr, fmt.Sprintf("Port 127.0.0.1:%d is already in use by intercept %s-1-%s", s.servicePort[1], s.Name(), s.AppNamespace()))
}
