import plotly
from plotly.offline import plot, init_notebook_mode

import plotly.graph_objs as go
import plotly.io as pio

import os
import numpy as np

import csv

# def parse_data_file(filename):
#   data = {}
#   with open(filename) as csv_file:
#     reader = csv.DictReader(csv_file, delimiter=',')
#     num_rows = 0
#     for row in reader:
#       if row['actualZone'] in data:
#         data[row['actualZone']] += 1
#       else:
#         data[row['actualZone']] = 0
#       num_rows += 1
#   return data

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
  for row in data:
    row = None
    for i in range(num_tasks):





csv_data = parse_data_file("../zoneResult.csv")
#colors = np.random.rand(N)

bar = [go.Bar(
  x=list(csv_data.keys()),
  y=list(csv_data.values())
)]

plot(bar)
