package main

import (
	"github.com/gradecak/fission-workflows/pkg/apiserver"
	"github.com/gradecak/fission-workflows/pkg/parse"
	"github.com/gradecak/fission-workflows/pkg/types"
	"github.com/gradecak/fission-workflows/pkg/types/typedvalues"
	"github.com/sirupsen/logrus"
	"math/rand"
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
	*apiserver.Client
}

func NewFWClient(url string) (*FWClient, error) {
	client, err := apiserver.Connect(url)
	if err != nil {
		return nil, err
	}
	return &FWClient{client}, nil
}

func (c FWClient) Invoke(ctx Context, wfID string) (*Result, error) {
	spec := types.NewWorkflowInvocationSpec(wfID, time.Now().Add(defaultTimeout))
	spec.Inputs = map[string]*typedvalues.TypedValue{}
	spec.ConsentId = "test" //RandomString(5)
	_, err := c.Invocation.InvokeSync(ctx, spec)
	if err != nil {
		return nil, err
	}
	return nil, nil
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
	md, err := c.Workflow.CreateSync(ctx, spec)
	if err != nil {
		return "", err
	}
	logrus.Info(md.ID())
	return md.ID(), nil
}
