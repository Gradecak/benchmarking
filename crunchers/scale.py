import pandas as pd
import numpy as np
import json
from pandas.io.json import json_normalize
import sys
from matplotlib.ticker import FormatStrFormatter

import matplotlib.pyplot as plt
import matplotlib.dates as mdates
from matplotlib.backends.backend_pdf import PdfPages

quantiles = [
    "quantiles.0",
	"quantiles.0.01",
	"quantiles.0.1",
	"quantiles.0.25",
	"quantiles.0.5",
	"quantiles.0.75",
	"quantiles.0.9",
]

linestyles = ['D-', 's-', 'v-', 'o-', '*-', 'x-']

plot_kwargs = {
    'style':linestyles,
    'markevery':7,
    'markersize':4
}

def plot_schedule_time(frame):
    df = frame.loc[(frame['name'] == 'workflows_scheduler_eval_time')]
    quantiles = "quantiles.0.5"
    df.loc[:,quantiles] = df.loc[:,quantiles].mul(1e-6)
    table = df.pivot_table(index="ThroughputBracket", values=quantiles, columns="ExpLabel")
    ax = table.plot(**plot_kwargs)
    handles, labels  = ax.get_legend_handles_labels()
    # handles = sorted(handles, key=lambda t: t.get_ydata(orig=True)[len(t.get_ydata(orig=True))-1], reverse=True)
    ax.legend(reversed(handles), [x[6:] for x in reversed(labels)], title="Policies")
    ax.set_ylim(ymin=0)
    ax.set_xlim(xmin=0)
    ax.set(xlabel="Queries Per Second (QPS)", ylabel="Scheduler Runtime (ms)")

def main():
    frame = pd.read_csv(sys.argv[1])
    if len(sys.argv) > 2 and sys.argv[2] == "pdf":
        with PdfPages('pdf-graphs/policy-scale.pdf') as pdf:
            plot_schedule_time(frame)
            pdf.savefig(bbox_inches='tight')
            plt.close()
    else:
        plot_schedule_time(frame)
        plt.show()


if __name__ == "__main__":
    main()
