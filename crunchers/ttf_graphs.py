import pandas as pd
import json
from pandas.io.json import json_normalize
import sys
from matplotlib.ticker import FormatStrFormatter

import matplotlib.pyplot as plt
import matplotlib.dates as mdates

def main():
    frame = pd.read_csv(sys.argv[1])
    pl1 = frame.pivot_table(index="PercentDFTasks", values="TimeToFailure", columns="NZones").plot()
    pl1.set(xlabel="% Multizone Tasks", ylabel="Time to Violation (ns)")
    pl2 = frame.pivot_table(index="PercentDFTasks", values="RunsBeforeFailure", columns="NZones").plot(kind='bar')
    # labels = ["0.{}".format(i+1) for i, item in enumerate(pl2.get_xticklabels())]
    pl2.set_xticklabels(["0.1", "0.2", "0.3", "0.4", "0.5", "0.6", "0.7", "0.8", "0.9", "1.0"])
    pl2.set(xlabel="% Multizone Tasks", ylabel="# Invocations before violation")
    plt.show()


if __name__ == "__main__":
    main()
