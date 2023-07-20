import pandas as pd
import matplotlib.pyplot as plt
from bidi.algorithm import get_display
import arabic_reshaper

def make_farsi_text(x):
    reshaped_text = arabic_reshaper.reshape(x)
    farsi_text = get_display(reshaped_text)
    return farsi_text

TEXT = '#e5e5e5'
MAIN = '#fca311'
plt.rcParams.update({
    'font.size': 8,
    'text.color': TEXT,
    'axes.labelcolor': TEXT,
    'axes.edgecolor': MAIN,
    'xtick.color': TEXT,
    'ytick.color': TEXT,
})


# Read the CSV file
data = pd.read_csv('output2.csv')

fig, ax = plt.subplots(1)

# xs = data.Time.values
# ys = data.DevicePerController.values
# ax.set_xlabel(make_farsi_text("Duration"))
# ax.set_ylabel(make_farsi_text("IoT Device Count For One Controller"))
# ax.set_title(make_farsi_text("IoT Device Count For One Controller To Duration"))

xs = data.Time.values
ys = data.Controllers.values
ax.set_xlabel(make_farsi_text("Duration"))
ax.set_ylabel(make_farsi_text("Controller Count"))
ax.set_title(make_farsi_text("Controllers Count To Durations"))
ax.plot(xs, ys, marker="o", color="#ffffff")
fig.autofmt_xdate(rotation=45)
plt.savefig('controllers_performance_benchmark.png', format='png', dpi=float(600), transparent=True)
plt.show()


