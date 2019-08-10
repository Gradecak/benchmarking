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

def combine_dfs(df_paths):
    li = []
    for path in df_paths:
            df = pd.read_json(path)
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
            li.append(df)
    frame = pd.concat(li, axis=0, ignore_index=True)
    # Cast strings to numerics
    # for qrange in quantile_range:
    #     frame.loc[:,qrange] = pd.to_numeric(frame[qrange] ,errors='coerce')
    # frame.loc[:,['value']] = pd.to_numeric(frame['value'] ,errors='coerce')
    # frame.loc[:,['sum']] = pd.to_numeric(frame['sum'] ,errors='coerce')
    return frame


def main():
    frame = combine_dfs(sys.argv[2:])
    print(frame)
    frame.to_csv(sys.argv[1])


if __name__ == "__main__":
    main()
