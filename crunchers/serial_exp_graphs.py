import pandas as pd
import json
from pandas.io.json import json_normalize
import sys
import matplotlib.pyplot as plt
import matplotlib.dates as mdates
from matplotlib.backends.backend_pdf import PdfPages

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

linestyles = ['D-', 's-', 'v-', 'o-', '*-', 'x-']

plot_kwargs = {
    'style':linestyles,
    'markevery':5,
    'markersize':4
}

def plot_average_throughput(frame):
    df = frame.loc[(frame['name'] == 'invocation_monitor_time')]
    df.loc[:,'quantiles.0.5'] = df.loc[:,'quantiles.0.5'].mul(1e-9)
    table = df.pivot_table(index="WfLength", columns="ExpLabel", values="quantiles.0.5")

    ax = table.plot(**plot_kwargs)
    handles, _  = ax.get_legend_handles_labels()
    handles = sorted(handles, key=lambda t: t.get_ydata(orig=True)[len(t.get_ydata(orig=True))-1])
    ax.set(xlabel="Task Count (Serial)", ylabel="Latency (s)")
    print(handles[0])
    ax.legend(reversed(handles),
              reversed(['base', 'prov', 'prov+consent', 'prov+consent(10ms)', 'prov-consent(50ms)', 'prov-consent(100ms)']),
              title="Treatments")

def plot_throughput_distributions(frame):
    exp_labels = ['base-throughput', 'prov-serial', 'prov-consent-serial', 'prov-consent-serial-10ms', 'prov-consent-serial-50ms', 'prov-consent-100ms-serial']
    f, axes = plt.subplots(2,3, sharey=True, sharex=True)
    i = 0
    for a in range(len(axes)):
        for b in range(len(axes[a])):
            label = exp_labels[i]
            df = frame.loc[(frame['name']=='invocation_monitor_time') & (frame['ExpLabel']== label)]
            df.loc[:,quantile_range] = df.loc[:,quantile_range].mul(1e-6)
            table = df.pivot_table(index="WfLength", columns="ExpLabel", values=quantile_range)
            axx = table.plot(ax=axes[a][b])
            i += 1
            axx.legend([], title=label)
            axx.set(xlabel="Task Count (Serial)", ylabel="Latency (ms)")

def main():
    frame = pd.read_csv(sys.argv[1])
    if len(sys.argv) > 2 and sys.argv[2] == "pdf":
        with PdfPages('serial_delay.pdf') as pdf:
            plot_average_throughput(frame)
            pdf.savefig(bbox_inches='tight')
            plt.close()
    else:
        plot_average_throughput(frame)
        plt.show()


if __name__ == "__main__":
    main()
