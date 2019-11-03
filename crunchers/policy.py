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

def plot_distributions(frame):
    pass
    # exp_labels = frame["ExpLabel"].unique()
    # f, axes = plt.subplots(1,2, sharey=True, sharex=True)
    # for idx, label in enumerate(exp_labels):
    #         df = frame.loc[(frame['name']=='policy_recovery_time') &
    #                        (frame['ExpLabel']== label)]
    #         df.loc[:,quantiles] = df.loc[:,quantiles].mul(1e-6)
    #         table = df.pivot_table(index="GraphSize", columns="ExpLabel", values=quantiles)
    #         axx = table.plot(ax=axes[idx], **plot_kwargs)
    #         axx.set_ylim(ymin=0)
    #         axx.set_xlim(xmin=10)
    #         axx.legend([], title=label)
    #         handles, _ = axes[0].get_legend_handles_labels()
    #         bb=(0.1,0.25)
    #         f.legend(reversed(handles),
    #              [x[10:] for x in reversed(quantiles)],
    #              loc=6,
    #                  bbox_to_anchor=bb,  prop={'size': 8})
    #         axx.set(xlabel="Queries Per Second (QPS)", ylabel="Latency (ms)")


def plot_schedule_time(frame):
    df = frame.loc[(frame['name'] == 'workflows_scheduler_eval_time')]
    quantiles = "quantiles.0.5"
    df.loc[:,quantiles] = df.loc[:,quantiles].mul(1e-9)
    table = df.pivot_table(index="ColdStart", values=quantiles, columns="ExpLabel")
    ax = table.plot(**plot_kwargs)
    # handles, _  = ax.get_legend_handles_labels()
    # handles = sorted(handles, key=lambda t: t.get_ydata(orig=True)[len(t.get_ydata(orig=True))-1], reverse=True)
    # ax.legend(handles, [x[10:] for x in reversed(quantiles)])
    # ax.set_ylim(ymin=0)
    # ax.set_xlim(xmin=10)

def plot_mean(frame):
    df = frame.loc[(frame['name'] == 'invocation_monitor_time') &
                         (frame['labels.type'] == "running")]
    df.loc[:,"quantiles.0.5"] = df.loc[:,"quantiles.0.5"].mul(1e-9)
    df.loc[:,"ColdStart"] = df.loc[:,"ColdStart"].mul(1e-9)
    table = df.pivot_table(index="ColdStart", values="quantiles.0.5", columns="ExpLabel")
    ax = table.plot(**plot_kwargs)
    # ax.set(xlabel="Queries Per Second (QPS)", ylabel="Latency (ms)")
    # ax.set_ylim(ymin=0)
    # ax.set_xlim(xmin=10)


def main():
    frame = pd.read_csv(sys.argv[1])
    if len(sys.argv) > 2 and sys.argv[2] == "pdf":
        with PdfPages('policy.pdf') as pdf:
            pdf.savefig(bbox_inches='tight')
            plt.close()
            plot_mean(frame)
            pdf.savefig(bbox_inches='tight')
            plt.close()
    else:
        plot_mean(frame)
        plot_schedule_time(frame)
        plt.show()


if __name__ == "__main__":
    main()
