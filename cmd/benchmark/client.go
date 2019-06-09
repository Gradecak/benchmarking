package main

import (
	"github.com/gradecak/fission-workflows/pkg/apiserver/httpclient"
	"github.com/gradecak/fission-workflows/pkg/parse"
	"github.com/gradecak/fission-workflows/pkg/types"
	"github.com/gradecak/fission-workflows/pkg/types/typedvalues"
	"github.com/sirupsen/logrus"
	"math/rand"
	"net/http"
	"os"
	"time"
)

const (
	defaultTimeout = 2 * time.Minute
)

type Result struct {
	response  string
	duration  time.Duration
	timestamp time.Time
	start     time.Time
}

func RandomString(n int) string {
	var letter = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	b := make([]rune, n)
	for i := range b {
		b[i] = letter[rand.Intn(len(letter))]
	}
	return string(b)
}

type FWClient struct {
	invocation *httpclient.InvocationAPI
	workflow   *httpclient.WorkflowAPI
}

func NewFWClient(url string) *FWClient {
	httpClient := http.Client{}
	return &FWClient{
		httpclient.NewInvocationAPI(url, httpClient),
		httpclient.NewWorkflowAPI(url, httpClient),
	}
}

func (c FWClient) Invoke(ctx Context, wfID string) (*Result, error) {
	spec := types.NewWorkflowInvocationSpec(wfID, time.Now().Add(defaultTimeout))
	// spec := types.NewWorkflowInvocationSpec(wfID)
	spec.Inputs = map[string]*typedvalues.TypedValue{}
	spec.ConsentId = "test" //RandomString(5)
	result := &Result{}
	start := time.Now()
	// ctx := context.TODO()
	wfi, err := c.invocation.InvokeSync(ctx, spec)
	result.timestamp = time.Now()
	if err != nil {
		return nil, err
	}
	result.duration = result.timestamp.Sub(start)
	wiStatus := wfi.GetStatus()
	if wiStatus.Successful() {
		result.response = typedvalues.MustUnwrap(wiStatus.GetOutput()).(string)
	}
	return result, nil
}

func (c FWClient) setupWF(ctx Context, specPath string) (string, error) {
	fd, err := os.Open(specPath)
	if err != nil {
		return "", err
	}
	spec, err := parse.Parse(fd)
	if err != nil {
		return "", err
	}
	md, err := c.workflow.CreateSync(ctx, spec)
	if err != nil {
		return "", err
	}
	logrus.Info(md.ID())
	return md.ID(), nil
}
