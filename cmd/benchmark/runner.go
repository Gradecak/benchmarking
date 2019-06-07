package main

import (
	"encoding/csv"
	"errors"
	"github.com/sirupsen/logrus"
	"os"
)

func Run(ctx Context, e Experiment, outFile string) error {
	res, err := e.Run(ctx)
	if err != nil {
		logrus.Fatal(err.Error())
		return err
	}
	return Save(res, outFile)
}

func Save(results [][]string, fileName string) error {
	if results == nil {
		logrus.Warn("Experiment concluded with empty results")
		return errors.New("Experiment Returned Empty Results")
	}
	logrus.Infof("Flat results: %+v \n", results)
	//open file
	file, err := os.Create(fileName)
	if err != nil {
		panic("Cannot open results file")
	}
	w := csv.NewWriter(file)
	defer file.Close()
	defer w.Flush()

	for _, r := range results {
		if err = w.Write(r); err != nil {
			return err
		}
	}
	return nil
}
