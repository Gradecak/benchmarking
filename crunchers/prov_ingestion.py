import pandas as pd
import json
from pandas.io.json import json_normalize
import sys
import matplotlib.pyplot as plt
import matplotlib.dates as mdates
from matplotlib.backends.backend_pdf import PdfPages


quantiles = [
    "quantiles.0.25",
    "quantiles.0.5",
    "quantiles.0.9",
    "quantiles.1",
]

quantile_labels = [
    "0.25",
    "0.5",
    "0.9",
    "1",
]

linestyles = ['-', '--', '-.', ':']

def plot_throughput(frame):
    df = frame.loc[(frame["name"] == 'dispatcher_total_time') &
                   (frame['labels.event'] == "PROVENANCE") &
                   (frame['QPS'] < 200)]
    df.loc[:,quantiles] = df.loc[:,quantiles].mul(1e-9)
    ax = df.pivot_table(index="QPS", values=quantiles).plot(style=linestyles)
    ax.set(xlabel="Queries Per Second (QPS)", ylabel="Latency (s)")
    ax.legend(quantile_labels, title="Qunatiles")


def plot_queue_time(frame):
    tframe = frame.loc[(frame["name"] == 'dispatcher_total_time') &
                            (frame['labels.event'] == "PROVENANCE") &
                            (frame['QPS'] < 320)].copy()
    eframe = frame.loc[(frame["name"] == 'dispatcher_enforce_time') &
                            (frame['labels.event'] == "PROVENANCE") &
                            (frame['QPS'] < 320)].copy()

    eframe.index = tframe.index

    # eframe = eframe.assign(total_time=tframe["quantiles.0.5"])

    eframe["queued"] = tframe['quantiles.0.5']
    print(eframe["queued"])
    ax = eframe.pivot_table(index="QPS", values=["queued",'quantiles.0.5'] ).plot(style=linestyles)


def main():
    frame = pd.read_csv(sys.argv[1])
    if len(sys.argv) > 2 and sys.argv[2] == 'pdf':
        with PdfPages('ingest.pdf') as pdf:
            plot_throughput(frame)
            pdf.savefig(bbox_inches='tight')
            plt.close()
            plot_queue_time(frame)
            pdf.savefig(bbox_inches='tight')
            plt.close()
    else:
        plot_throughput(frame)
        plot_queue_time(frame)
        plt.tight_layout()
        plt.show()

if __name__ == "__main__":
    main()
