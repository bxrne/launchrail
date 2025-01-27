#!/usr/bin/env python3

import pandas as pd
import matplotlib.pyplot as plt
from mpl_toolkits.mplot3d import Axes3D


def plot_flight_data(csv_path):
    # Read data
    df = pd.read_csv(csv_path)

    # Create 3D plots
    fig = plt.figure(figsize=(15, 5))

    # Position plot
    ax1 = fig.add_subplot(131, projection="3d")
    ax1.plot(df["Sx"], df["Sy"], df["Sz"])
    ax1.set_title("Position")
    ax1.set_xlabel("X (m)")
    ax1.set_ylabel("Y (m)")
    ax1.set_zlabel("Z (m)")

    # Velocity plot
    ax2 = fig.add_subplot(132, projection="3d")
    ax2.plot(df["Vx"], df["Vy"], df["Vz"])
    ax2.set_title("Velocity")
    ax2.set_xlabel("X (m/s)")
    ax2.set_ylabel("Y (m/s)")
    ax2.set_zlabel("Z (m/s)")

    # Acceleration plot
    ax3 = fig.add_subplot(133, projection="3d")
    ax3.plot(df["Ax"], df["Ay"], df["Az"])
    ax3.set_title("Acceleration")
    ax3.set_xlabel("X (m/s²)")
    ax3.set_ylabel("Y (m/s²)")
    ax3.set_zlabel("Z (m/s²)")

    plt.tight_layout()
    plt.show()


if __name__ == "__main__":
    plot_flight_data("/Users/adambyrne/.launchrail/motion/2025-01-26T21:47:39.csv")
