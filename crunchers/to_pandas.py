import pandas as pd
import json
from pandas.io.json import json_normalize
import sys
import matplotlib.pyplot as plt

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



def plot_throughput(df):
    throughput_df = df.loc[(df['name'] == 'invocation_monitor_time')].copy()
    throughput_df.loc[:,quantile_range] = throughput_df.loc[:,quantile_range].mul(1e-6)
    # qr = quantile_range + ["ThroughputBracket"]
    # _, ax = plt.subplots()

    ax = throughput_df.pivot_table(index="ThroughputBracket", values=quantile_range).plot(kind='line')
    # ax2 = ax.twinx()
    p = throughput_df.pivot_table(index="ThroughputBracket", columns="labels.type", values="quantiles.0.5")
    p['running'] = p['running'].sub(p['queued'])
    p.plot(kind='bar', stacked=True)

    # ax.set(xlabel="QPS", ylabel="Latency (ms)")
    # ax.legend([x[10:] for x in quantile_range])

def plot_mem(df):
    mem_df = df.loc[(df['name'] == "go_memstats_alloc_bytes")].copy()
    ax = mem_df.groupby("ThroughputBracket")['value'].median().plot(kind='area')
    ax.set(xlabel="QPS", ylabel="Memory Consumed (Bytes)")

def plot_scheduler_eval(df):
    qps_brack = ['quantiles.0.5', 'quantiles.0.9']
    eval_df = df.loc[(df['name'] == 'workflows_scheduler_eval_time')].copy()
    ax = eval_df.groupby("ThroughputBracket")[qps_brack].mean().plot(kind='line')
    ax.set(xlabel="QPS", ylabel="Latency (ms)")
    ax.legend([x[10:] for x in qps_brack])

def plot_network_usage(df):
    net_df = df.loc[(df['name'] == "container_network_transmit_bytes_total") &
                    (df['labels.image'] == "mgradecak/fission-workflows-bundle:0.35") &
                    (df['labels.interface'] == "enp5s0")]
    ax = net_df.groupby("ThroughputBracket")['value'].mean().plot(kind='line')
    ax.set(xlabel="QPS", ylabel="Network Utilisation (Bytes)")
    # ax.legend([x[10:] for x in qps_brack])

def plot_active_controllers(df):
    ac_df = df.loc[(df['name'] == "system_controller_concurrent") &
                   (df['labels.system'] == "invocation")].copy()
    ax = ac_df.groupby('Timestamp')['value', 'ThroughputBracket'].mean().plot(kind='line')
    ax.axhline(1500, color='k', linestyle='--')


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

    # plt.hold(True)
    # plot_mem(df)
    plot_throughput(df)
    # plot_scheduler_eval(df)
    plot_active_controllers(df)
    # plot_network_usage(df)

    plt.show()

if __name__ == "__main__":
    main()
# print(df1.loc[0,:])
