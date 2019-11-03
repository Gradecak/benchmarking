import pandas as pd
import json
from pandas.io.json import json_normalize
import sys
import matplotlib.pyplot as plt
import matplotlib.dates as mdates
from matplotlib.backends.backend_pdf import PdfPages

quantile_range = [
    "quantiles.0.01",
    "quantiles.0.1",
    "quantiles.0.25",
    "quantiles.0.5",
    "quantiles.0.75",
    "quantiles.0.9",
]

ranges = [
    100,
    300,
    500,
    700
]

linestyles = ['D-', 's-', 'v-', 'o-', '*-', 'x-']

def plot_slowdown(frame):
    filtered = frame.loc[(frame['name'] == 'invocation_monitor_time') &
                         (frame['labels.type'] == "running")]
    filtered.loc[:,'quantiles.0.5'] = filtered.loc[:,'quantiles.0.5'].mul(1e-9)

    filtered = filtered.pivot_table(index="ThroughputBracket", values=["quantiles.0.5"], columns="ExpLabel")
    filtered["slowdown-prov"] = ((filtered["quantiles.0.5"]["prov-throughput"] - filtered["quantiles.0.5"]["base-throughput"])
                                 + filtered["quantiles.0.5"]["base-throughput"]) / filtered["quantiles.0.5"]["base-throughput"]
    filtered["slowdown-prov-consent"] = ((filtered["quantiles.0.5"]["prov-consent-throughput"] - filtered["quantiles.0.5"]["base-throughput"])
                                 + filtered["quantiles.0.5"]["base-throughput"]) / filtered["quantiles.0.5"]["base-throughput"]
    ax = filtered.pivot_table(index="ThroughputBracket", values=["slowdown-prov", "slowdown-prov-consent"]).plot(style=linestyles, markevery=7)
    ax.legend(['Prov', 'Prov+Cosnsent'], title="Extensions")


def plot_average_throughput(frame):
    df = frame.loc[(frame['name'] == 'invocation_monitor_time') &
                   (frame['labels.type'] == "running")]

    df.loc[:,'quantiles.0.5'] = df.loc[:,'quantiles.0.5'].mul(1e-6)
    table = df.pivot_table(index="ThroughputBracket", columns="ExpLabel", values="quantiles.0.5")

    ax = table.plot(style=linestyles, markevery=5, markersize=4)
    ax.set(xlabel="Queries Per Second (QPS)", ylabel="Latency (ms)")
    ax.set_ylim(ymin=0)
    ax.set_xlim(xmin=10)
    ax.legend(['Base', 'Prov+Cosnsent', 'Prov'], title="Extensions")

def plot_throughput_distributions(frame):
    exp_labels = ['base-throughput', 'prov-throughput', 'prov-consent-throughput']
    f, axes = plt.subplots(1,3, sharey=True, sharex=True)
    for idx, label in enumerate(exp_labels):
        df = frame.loc[(frame['name']=='invocation_monitor_time')
                       & (frame['ExpLabel']== label)
                       & (frame['labels.type'] == "running")]
        df.loc[:,quantile_range] = df.loc[:,quantile_range].mul(1e-6)
        table = df.pivot_table(index="ThroughputBracket", columns="ExpLabel", values=quantile_range)
        ax = table.plot(ax=axes[idx], style=linestyles, markevery=7, markersize=2, linewidth=1.0)
        ax.set(xlabel="", ylabel="Latency (ms)")
        ax.set_xticks([100,300,500,700])
        ax.legend([], title=label.replace("-throughput", ""))
        ax.tick_params(axis='x')
        f.text(0.5, 0.04, 'Queries Per Second (QPS)', ha='center', va='center')
        handles, _ = axes[0].get_legend_handles_labels()
        bb=(0.125,0.65)
        f.legend(reversed(handles),
                 [x[10:] for x in reversed(quantile_range)],
                 loc=6,
                 bbox_to_anchor=bb)


def box_plot(frame):
    labels = frame["ExpLabel"].unique()
    f, axes = plt.subplots(1,len(labels), sharey=True, sharex=True)
    for idx, label in enumerate(labels):
        df = frame.loc[(frame["name"] == 'invocation_monitor_time') &
                       (frame["ThroughputBracket"].isin(ranges)) &
                       (frame['ExpLabel']== label) &
                       (frame["labels.type"] == "running") ]
        df.loc[:,quantile_range] = df.loc[:,quantile_range].mul(1e-6)
        a = pd.melt(df, id_vars=['ThroughputBracket'], value_vars=quantile_range)
        d = a.pivot(columns="ThroughputBracket", values="value")
        ax = d.plot.box(ax=axes[idx])
        ax.legend([], title=label.replace("-throughput", ""))
        ax.set(xlabel="", ylabel="Latency (ms)")
        f.text(0.5, 0.04, 'Queries Per Second (QPS)', ha='center', va='center')


def plot_mem_alloc(frame):
    df = frame.loc[frame['name']=='go_memstats_alloc_bytes']
    df.loc[:,'value'] = df.loc[:,'value'].mul(1e-9)
    table = df.pivot_table(index="ThroughputBracket", columns="ExpLabel", values="value")
    ax = table.plot(style=linestyles, markevery=5, markersize=4)
    ax.set(xlabel="Queries Per Second (QPS)", ylabel="Memory Allocated (Gb)")
    ax.legend(['Base', 'Prov+Cosnsent', 'Prov'], title="Extensions")

def plot_network_usage(frame):
    df = frame.loc[(frame['name'] == "container_network_transmit_bytes_total") &
                   (frame['labels.image'].str.contains("mgradecak/fission-workflows-bundle"))
    ]
    df.loc[:,'value'] = df.loc[:,'value'].mul(1e-9)
    table = df.pivot_table(index="ThroughputBracket", columns="ExpLabel", values="value")
    ax = table.plot(style=linestyles, markevery=5, markersize=4)
    ax.set(xlabel="Time (Hrs)", ylabel="Transmitted (Gb)")
    ax.set_xticklabels([0, 0.5,1,1.5,2,2.5,3,3.5])
    ax.legend(['Base', 'Prov+Consent', 'Prov'], title="Extensions")

def main():
    frame = pd.read_csv(sys.argv[1])
    if len(sys.argv) > 2 and sys.argv[2] == "pdf":
        with PdfPages('throughput.pdf') as pdf:
            plot_average_throughput(frame)
            pdf.savefig(bbox_inches='tight')
            plt.close()
            plot_mem_alloc(frame)
            pdf.savefig(bbox_inches='tight')
            plt.close()
            plot_network_usage(frame)
            pdf.savefig(bbox_inches='tight')
            plt.close()
            box_plot(frame)
            pdf.savefig(bbox_inches='tight')
            plt.close()
            plot_throughput_distributions(frame)
            pdf.savefig(bbox_inches='tight')
            plt.close()
            plot_slowdown(frame)
            pdf.savefig(bbox_inches='tight')
            plt.close()
    else:
        # plot_average_throughput(frame)
        # plot_mem_alloc(frame)
        # plot_network_usage(frame)
        # box_plot(frame)
        # plot_throughput_distributions(frame)
        plot_slowdown(frame)
        plt.show()

if __name__ == "__main__":
    main()
