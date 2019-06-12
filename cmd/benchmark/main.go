package main

import (
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"os"
	"time"
)

func main() {
	app := cli.NewApp()
	app.Name = "boom"
	app.Usage = "make an explosive entrance"
	app.Flags = []cli.Flag{
		cli.IntFlag{
			Name:  "verbosity",
			Value: 2,
			Usage: "CLI verbosity (0 is quiet, 1 is the default, 2 is verbose.)",
		},
	}
	app.Action = commandContext(func(ctx Context) error {
		logrus.Info("Parsing Experiment Config...")
		conf, err := ConfigFromFile(ctx.Args().Get(0))
		logrus.Infof("%+v", conf.CollectorConf)
		if err != nil {
			logrus.Error(err.Error())
			return err
		}
		exp, err := GetExperiment(conf)
		if err != nil {
			logrus.Fatal(err)
			return err
		}
		logrus.Info("Setting up Experiment Config...")
		return Run(ctx, exp, conf.OutputFile)
	})
	app.Run(os.Args)
}

type Context struct {
	*cli.Context
}

func (c Context) Deadline() (deadline time.Time, ok bool) {
	return
}

func (c Context) Done() <-chan struct{} {
	return nil
}

func (c Context) Err() error {
	return nil
}

func (c Context) Value(key interface{}) interface{} {
	if s, ok := key.(string); ok {
		return c.Generic(s)
	}
	return nil
}

func commandContext(fn func(c Context) error) func(c *cli.Context) error {
	return func(c *cli.Context) error {
		switch c.GlobalInt("verbosity") {
		case 0:
			logrus.SetLevel(logrus.ErrorLevel)
		case 1:
			logrus.SetLevel(logrus.DebugLevel)
		default:
			fallthrough
		case 2:
			logrus.SetLevel(logrus.InfoLevel)
		}
		return fn(Context{c})
	}
}
