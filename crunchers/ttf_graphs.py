import pandas as pd
import json
from pandas.io.json import json_normalize
import sys
from matplotlib.ticker import FormatStrFormatter

import matplotlib.pyplot as plt
import matplotlib.dates as mdates
from matplotlib.backends.backend_pdf import PdfPages

linestyles = ['D-', 's-', 'v-', 'o-', '*-', 'x-']

plot_kwargs = {
    'style':linestyles,
    'markevery':1,
    'markersize':4
}

def plot_1(frame):
    pl1 = frame.pivot_table(index="PercentDFTasks", values="TimeToFailure", columns="NZones").plot(**plot_kwargs)
    pl1.set(xlabel="% Multizone Tasks", ylabel="Time to Violation (ns)")
    pl1.legend(title='# Available Zones')

def bar_plot(frame):
    pl2 = frame.pivot_table(index="PercentDFTasks", values="RunsBeforeFailure", columns="NZones").plot(kind='bar')
    pl2.legend(loc=1, ncol=4, title="# Available Zones")
    pl2.set_xticklabels(["0.1", "0.2", "0.3", "0.4", "0.5", "0.6", "0.7", "0.8", "0.9", "1.0"])
    pl2.set(xlabel="% Multizone Tasks", ylabel="# Invocations before violation")


def main():
    frame = pd.read_csv(sys.argv[1])
    if len(sys.argv) > 2 and sys.argv[2] == "pdf":
        with PdfPages('ttf.pdf') as pdf:
            plot_1(frame)
            pdf.savefig(bbox_inches='tight')
            plt.close()
            bar_plot(frame)
            pdf.savefig(bbox_inches='tight')
            plt.close()
    else:
        plot_1(frame)
        bar_plot(frame)
        plt.show()





if __name__ == "__main__":
    main()
