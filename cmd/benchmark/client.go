package main

import (
	"context"
	"math/rand"
	"os"
	"time"

	"github.com/gradecak/fission-workflows/pkg/apiserver"
	"github.com/gradecak/fission-workflows/pkg/parse/yaml"
	"github.com/gradecak/fission-workflows/pkg/types"
	"github.com/gradecak/fission-workflows/pkg/types/typedvalues"
	"github.com/sirupsen/logrus"
)

const (
	defaultTimeout = 3 * time.Minute
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

func (c FWClient) InvokeWithConsentID(ctx Context, wfID string, cId string) (*Result, error) {
	spec := types.NewWorkflowInvocationSpec(wfID, time.Now().Add(defaultTimeout))
	spec.Inputs = map[string]*typedvalues.TypedValue{}
	spec.ConsentId = cId
	_, err := c.Invocation.InvokeSync(ctx, spec)
	if err != nil {
		logrus.Errorf("Invocation Error %v", err)
		return nil, err
	}
	return nil, nil
}

func (c FWClient) Invoke(ctx Context, wfID string) (*Result, error) {
	return c.InvokeWithConsentID(ctx, wfID, RandomString(6))
	// return c.InvokeWithConsentID(ctx, wfID, "")
}

func (c FWClient) SetupWfFromFile(ctx Context, specPath string) (string, error) {
	fd, err := os.Open(specPath)
	if err != nil {
		return "", err
	}
	spec, err := yaml.Parse(fd)
	if err != nil {
		return "", err
	}
	md, err := c.Workflow.CreateSync(context.TODO(), spec)
	if err != nil {
		return "", err
	}
	logrus.Info(md.ID())
	return md.ID(), nil
}

func (c FWClient) SetupWfFromSpec(ctx Context, spec *types.WorkflowSpec) (string, error) {
	md, err := c.Workflow.CreateSync(context.TODO(), spec)
	if err != nil {
		return "", err
	}
	logrus.Infof("Created Workflow (%s)", md.ID())
	return md.ID(), nil
}

// func (c *Consent) Update(ctx context.Context, cm *types.ConsentMessage) (*empty.Empty, error) {
func (c FWClient) RevokeConsent(ctx Context, id string) error {
	cm := &types.ConsentMessage{
		ID:     id,
		Status: &types.ConsentStatus{types.ConsentStatus_REVOKED},
	}
	_, err := c.Consent.Update(ctx, cm)
	return err
}
