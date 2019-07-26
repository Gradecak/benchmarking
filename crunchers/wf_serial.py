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

def plot_latency(df):
    df1 = df.loc[(df['name'] == 'invocation_monitor_time')].copy()
    df1.loc[:,quantile_range] = df1.loc[:,quantile_range].mul(1e-6)

    ax = df1.loc[df1['labels.type']=='running'].pivot_table(index="WfLength", values=quantile_range).plot(kind="line")

def plot_throughput(df):
    throughput_df = df.loc[(df['name'] == 'invocation_monitor_time')].copy()
    throughput_df.loc[:,quantile_range] = throughput_df.loc[:,quantile_range].mul(1e-6)
    # qr = quantile_range + ["ThroughputBracket"]
    # _, ax = plt.subplots()

    ax = throughput_df.loc[throughput_df['labels.type'] == 'running'].pivot_table(index="Timestamp", columns="labels.type", values='quantiles.0.9')["running"].plot(kind='line')
    ax.xaxis.set_major_formatter(mdates.DateFormatter('%H:%M:%S'))
    k_df = df.loc[(df['name'] == "fes_cache_current_cache_counts") &
                  (df['labels.name'] == "invocation")].copy()
    k_df.groupby('Timestamp')['value'].mean().plot(kind='line', ax=ax)
    periods = throughput_df[throughput_df.ThroughputBracket.diff()!=0].Timestamp.values
    # Plot the red vertical lines
    for item in periods[1::]:
        plt.axvline(item, ymin=0, ymax=1,color='red', linestyle='-.')
    # ax2 = ax.twinx()
    p = throughput_df.pivot_table(index="ThroughputBracket", columns="labels.type", values="quantiles.0.5")
    p['running'] = p['running'].sub(p['queued'])
    p.plot(kind='bar', stacked=True)


    throughput_df.loc[throughput_df['labels.type']=="running"].pivot_table(index="ThroughputBracket", values=quantile_range).plot(kind='line')
    # ax.set(xlabel="QPS", ylabel="Latency (ms)")
    # ax.legend([x[10:] for x in quantile_range])




def main():
    df = pd.read_json(sys.argv[1])
    # flatten the State column
    df = (pd.concat({i: json_normalize(x) for i, x in df.pop('State').items()}, sort=False)
          .reset_index(level=1, drop=True)
          .join(df)
          .reset_index(drop=True))
    # Flatten the metrics colun
    df = (pd.concat({i: json_normalize(x) for i, x in df.pop('metrics').items()}, sort=False)
          .reset_index(level=1, drop=True)
          .join(df)
          .reset_index(drop=True))

    # Cast strings to numerics
    for qrange in quantile_range:
        df.loc[:,qrange] = pd.to_numeric(df[qrange] ,errors='coerce')
    df.loc[:,['value']] = pd.to_numeric(df['value'] ,errors='coerce')
    df.loc[:,['sum']] = pd.to_numeric(df['sum'] ,errors='coerce')

    plot_latency(df)

    plt.show()

if __name__ == "__main__":
    main()
