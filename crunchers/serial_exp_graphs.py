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
    df.loc[:,'quantiles.0.5'] = df.loc[:,'quantiles.0.5'].mul(1e-9)
    table = df.pivot_table(index="WfLength", columns="ExpLabel", values="quantiles.0.5")

    ax = table.plot()
    ax.set(xlabel="Task Count (Serial)", ylabel="Latency (s)")
    # ax.legend(['base', 'prov+consent(100ms)', 'prov+consent', 'prov+consent(10ms)', 'prov-consent(50ms)', 'prov'], title="Treatments")

# def plot_throughput_distributions(frame):
#     exp_labels = ['base-throughput', 'prov-serial', 'prov-consent-serial', 'prov-consent-serial-10ms', 'prov-consent-serial-50ms', 'prov-consent-100ms-serial']
#     f, axes = plt.subplots(2,3, sharey=True)
#     for idx, label in enumerate(exp_labels):
#         df = frame.loc[(frame['name']=='invocation_monitor_time') & (frame['ExpLabel']== label)]
#         df.loc[:,quantile_range] = df.loc[:,quantile_range].mul(1e-6)
#         table = df.pivot_table(index="WfLength", columns="ExpLabel", values=quantile_range)
#         i = int(idx / 3)
#         j = int(idx*i + 1)
#         print("i {} j {} idx {}".format(i,j, idx))
#         ax = table.plot(ax=axes[i][j])
#         # ax.legend([], title=label)
#         # ax.set(xlabel="Queries Per Second (QPS)", ylabel="Latency (ms)")
#     # s = f.subplotpars
#     # bb=[s.left, s.top+0.01, s.right-s.left, 0.03 ]
#     # f.legend([x[10:] for x in quantile_range],
#     #          loc=8,
#     #          title="Quantiles",
#     #          mode="expand",
#     #          bbox_to_anchor=bb,
#    #          ncol=len(quantile_range),
#     #          bbox_transform=f.transFigure,
#     #          fancybox=False, edgecolor="k")

def main():
    frame = pd.read_csv(sys.argv[1])
    plot_average_throughput(frame)
    # plot_throughput_distributions(frame)
    plt.show()

if __name__ == "__main__":
    main()
