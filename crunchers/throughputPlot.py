import plotly
from plotly.offline import plot, init_notebook_mode
import plotly.graph_objs as go
import plotly.io as pio


import sys
import numpy as np
import csv
from datetime import datetime

TO_SEC = 10**-9
quantiles = [0, 0.1, 0.25, 0.5, 0.75, 0.9, 1]

def parse_data_file(filename):
  data = {}
  with open(filename) as csv_file:
    reader = csv.DictReader(csv_file, delimiter=',')
    for row in reader:
        if row["qps"] in data:
            data[row["qps"]].append(int(row["latency"]) * TO_SEC)
        else:
            data[row["qps"]] = [int(row["latency"]) * TO_SEC]
  return data

def crunch(data):
    throughputs = []
    percentile_data = {}
    traces = []
    for k, v in data.items():
        throughputs.append(int(k))
        percentile_data[int(k)] = np.quantile(v, quantiles)

    for i in range(len(quantiles)):
        qrange = []
        for t in throughputs:
            qrange.append(percentile_data[t][i])

        traces.append(
            go.Scatter(
                x = throughputs,
                y = qrange,
                name = quantiles[i]
            )
        )

    layout = dict(title = 'Test',
                  xaxis = dict(title = 'QPS'),
                  yaxis = dict(title = 'Latency'),
              )

    fig = dict(data=traces, layout=layout)
    plot(fig, filename='styled-line')




if __name__ == "__main__":
    data = parse_data_file(sys.argv[1])
    crunch(data)
