import pandas as pd
import matplotlib.pyplot as plt
from bidi.algorithm import get_display
import arabic_reshaper

def make_farsi_text(x):
    reshaped_text = arabic_reshaper.reshape(x)
    farsi_text = get_display(reshaped_text)
    return farsi_text

plt.rcParams.update({'font.size': 16})

# Read the CSV file
data = pd.read_csv('output.csv')


fig, ax = plt.subplots(1)

xs = data.Time.values
ys = data.DevicePerController.values
ax.set_xlabel(make_farsi_text("مدت زمان"))
ax.set_ylabel(make_farsi_text("تعداد دستگاه به ازای یک کنترل‌کننده"))
ax.set_title(make_farsi_text("تعداد دستگاه به ازای یک کنترل‌کننده به مدت زمان"))

# xs = data.Time.values
# ys = data.Controllers.values
# ax.set_xlabel(make_farsi_text("مدت زمان"))
# ax.set_ylabel(make_farsi_text("تعداد کنترل‌کننده"))
# ax.set_title(make_farsi_text("تعداد کنترل‌کننده به مدت زمان"))
ax.plot(xs, ys, marker="o")
fig.autofmt_xdate(rotation=45)
plt.show()


