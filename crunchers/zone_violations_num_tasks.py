import plotly
from plotly.offline import plot, init_notebook_mode


import plotly.graph_objs as go
import plotly.io as pio

import os
import numpy as np

from functools import reduce

import csv

def parse_data_file(filename):
  data = {}
  with open(filename) as csv_file:
    reader = csv.DictReader(csv_file, delimiter=',')
    num_rows = 0
    for row in reader:
      if row['T1'] in data:
        data[row['T1']] += 1
      else:
        data[row['T1']] = 0
      num_rows += 1
  return data

def read_data_file(filename):
  data = []
  with open(filename) as csv_file:
    reader = csv.DictReader(csv_file, delimiter=',')
    num_rows = 0
    for row in reader:
      data.append(row)
      num_rows += 1
  return (data, num_rows)

def crunch(data, num_tasks, num_rows):
  expected_zone = data[0]['expectedZone']
  task_results = []

  success = []
  for row in data:
    s_row = []
    for i in range(num_tasks):
      s = 1 # we assume as we start that the zones are matching
      for j in range(0, i+1):
        if s != 0:
          key = "T{}".format(j+1)
          # if the targetZone and actualZone match and previous zone comparisons
          # have succeeded
          if row["expectedZone"] != row[key]:
            s = 0
      s_row.append(s)
    success.append(tuple(s_row))
  print(success)

  data = reduce(lambda x,y: tuple(map(sum, zip(x,y))), success)
  print("total Rows {} --- {}".format(num_rows, data))

  successTrace = go.Bar(
    x=["1-Task", "2-Tasks", "3-Tasks"],
    y=data,
    name="Success"
  )

  failTrace = go.Bar(
    x=["1-Task", "2-Tasks", "3-Tasks"],
    y = [num_rows-data[0], num_rows-data[1], num_rows-data[2]],
    name = "Failures"
  )

  data = [successTrace, failTrace]
  layout = go.Layout(
    barmode='stack'
  )

  fig = go.Figure(data=data, layout=layout)
  plot(fig, filename='stacked-bar')



#trace 1 succeeded
#trace 2 failed


csv_data, num_rows = read_data_file("../zoneResult.csv")
crunch(csv_data, 3, num_rows)

csv_data =  parse_data_file("../zoneResult.csv")
#colors = np.random.rand(N)

bar = [go.Bar(
  x=list(csv_data.keys()),
  y=list(csv_data.values())
)]

plot(bar)
