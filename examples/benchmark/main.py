import csv
import json

# json_data = '''
# [
#     {"results":{"Time":"320.927926ms","TotalControllers":1,"TotalDevices":1},"settings":{"Controllers":1,"DevicesPerController":1,"Duration":"5s"}},
#     {"results":{"Time":"325.609733ms","TotalControllers":1,"TotalDevices":10},"settings":{"Controllers":1,"DevicesPerController":10,"Duration":"5s"}},
#     {"results":{"Time":"328.67701ms","TotalControllers":1,"TotalDevices":100},"settings":{"Controllers":1,"DevicesPerController":100,"Duration":"5s"}},
#     {"results":{"Time":"332.549921ms","TotalControllers":1,"TotalDevices":1000},"settings":{"Controllers":1,"DevicesPerController":1000,"Duration":"5s"}},
#     {"results":{"Time":"403.699692ms","TotalControllers":1,"TotalDevices":10000},"settings":{"Controllers":1,"DevicesPerController":10000,"Duration":"5s"}},
#     {"results":{"Time":"1.150619205s","TotalControllers":1,"TotalDevices":100000},"settings":{"Controllers":1,"DevicesPerController":100000,"Duration":"5s"}}
# ]
# '''

json_data = '''
[
    {"results":{"Time":"316.631433ms","TotalControllers":1,"TotalDevices":1},"settings":{"Controllers":1,"DevicesPerController":1,"Duration":"5s"}},
    {"results":{"Time":"340.973697ms","TotalControllers":10,"TotalDevices":10},"settings":{"Controllers":10,"DevicesPerController":1,"Duration":"5s"}},
    {"results":{"Time":"632.569765ms","TotalControllers":100,"TotalDevices":100},"settings":{"Controllers":100,"DevicesPerController":1,"Duration":"5s"}},
    {"results":{"Time":"3.487017751s","TotalControllers":1000,"TotalDevices":1000},"settings":{"Controllers":1000,"DevicesPerController":1,"Duration":"5s"}},
    {"results":{"Time":"34.187790534s","TotalControllers":10000,"TotalDevices":10000},"settings":{"Controllers":10000,"DevicesPerController":1,"Duration":"1m0s"}},
    {"results":{"Time":"45.409622456s","TotalControllers":100000,"TotalDevices":100000},"settings":{"Controllers":100000,"DevicesPerController":1,"Duration":"2m0s"}}
]
'''

data = json.loads(json_data)

# Prepare CSV file
csv_file = open('output2.csv', 'w', newline='')
csv_writer = csv.writer(csv_file)
csv_writer.writerow(['Controllers', 'Time', 'Description'])

# Convert JSON to CSV rows
for item in data:
    device_per_controller = item['settings']['Controllers']
    time = item['results']['Time']
    description = ', '.join([f'{k}: {v}' for k, v in item['settings'].items()])
    csv_writer.writerow([device_per_controller, time, description])

# Close CSV file
csv_file.close()

