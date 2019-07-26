import pandas as pd
import json
from pandas.io.json import json_normalize
import sys
import matplotlib.pyplot as plt
import matplotlib.dates as mdates

quantile_range = [
    # "quantiles.0",
    # "quantiles.0.001",
    "quantiles.0.01",
    # "quantiles.0.02",
    "quantiles.0.1",
    # "quantiles.0.2",
    "quantiles.0.25",
    "quantiles.0.5",
    "quantiles.0.75",
    "quantiles.0.9",
    # "quantiles.0.98",
    # "quantiles.0.99",
    # "quantiles.0.999",
    # "quantiles.0.1"
]

def plot_average_throughput(frame):
    df = frame.loc[(frame['name'] == 'invocation_monitor_time')]
    df.loc[:,'quantiles.0.5'] = df.loc[:,'quantiles.0.5'].mul(1e-6)
    table = df.pivot_table(index="ThroughputBracket", columns="ExpLabel", values="quantiles.0.5")

    ax = table.plot()
    ax.set(xlabel="Queries Per Second (QPS)", ylabel="Latency (ms)")
    ax.legend(['Base', 'Prov+Cosnsent', 'Prov'], title="Extensions")


def plot_throughput_distributions(frame):
    exp_labels = ['base-throughput', 'prov-throughput', 'prov-consent-throughput']
    f, axes = plt.subplots(1,3, sharey=True)
    for idx, label in enumerate(exp_labels):
        df = frame.loc[(frame['name']=='invocation_monitor_time') & (frame['ExpLabel']== label)]
        df.loc[:,quantile_range] = df.loc[:,quantile_range].mul(1e-6)
        table = df.pivot_table(index="ThroughputBracket", columns="ExpLabel", values=quantile_range)
        ax = table.plot(ax=axes[idx])
        ax.legend([], title=label)
        ax.set(xlabel="Queries Per Second (QPS)", ylabel="Latency (ms)")
    s = f.subplotpars
    bb=[s.left, s.top+0.01, s.right-s.left, 0.03 ]
    f.legend([x[10:] for x in quantile_range],
             loc=8,
             title="Quantiles",
             mode="expand",
             bbox_to_anchor=bb,
             ncol=len(quantile_range),
             bbox_transform=f.transFigure,
             fancybox=False, edgecolor="k")


def plot_mem_alloc(frame):
    df = frame.loc[frame['name']=='go_memstats_alloc_bytes']
    df.loc[:,'value'] = df.loc[:,'value'].mul(1e-9)
    table = df.pivot_table(index="ThroughputBracket", columns="ExpLabel", values="value")
    ax = table.plot()
    ax.set(xlabel="Queries Per Second (QPS)", ylabel="Memory Allocated (Gb)")
    ax.legend(['Base', 'Prov+Cosnsent', 'Prov'], title="Extensions")

def plot_network_usage(frame):
    df = frame.loc[(frame['name'] == "container_network_transmit_bytes_total") &
                    (frame['labels.image'].str.contains("mgradecak/fission-workflows-bundle"))
                    ]
    df.loc[:,'value'] = df.loc[:,'value'].mul(1e-9)
    table = df.pivot_table(index="ThroughputBracket", columns="ExpLabel", values="value")
    ax = table.plot()
    ax.set(xlabel="Queries Per Second (QPS)", ylabel="Transmitted (Gb)")
    ax.legend(['Base', 'Prov+Cosnsent', 'Prov'], title="Extensions")

def main():
    frame = pd.read_csv(sys.argv[1])
    plot_average_throughput(frame)
    plot_mem_alloc(frame)
    plot_network_usage(frame)
    plot_throughput_distributions(frame)
    plt.show()

if __name__ == "__main__":
    main()
