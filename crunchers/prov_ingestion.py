import pandas as pd
import json
from pandas.io.json import json_normalize
import sys
import matplotlib.pyplot as plt
import matplotlib.dates as mdates

quantiles = [
    "quantiles.0.25",
    "quantiles.0.5",
    "quantiles.0.9",
    "quantiles.1",
]

def plot_throughput(frame):
    df = frame.loc[(frame["name"] == 'dispatcher_enforcment_time') &
                   (frame['labels.event'] == "PROVENANCE")]
    ax = df.pivot_table(index="QPS", values=quantiles).plot()


def main():
    frame = pd.read_csv(sys.argv[1])
    plot_throughput(frame)
    plt.show()

if __name__ == "__main__":
    main()
