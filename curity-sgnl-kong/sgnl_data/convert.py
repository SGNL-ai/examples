#Simple Python Script to convert records.json to CSV format needed to import patients
# into the SGNL graph.

import json
import csv

with open('records.json') as json_file:
	jsondata = json.load(json_file)

data_file = open('record_metadata.csv', 'w', newline='')
csv_writer = csv.writer(data_file)

count = 0
for data in jsondata:
	if count == 0:
		header = data.keys()
		csv_writer.writerow(header)
		count += 1
	csv_writer.writerow(data.values())

data_file.close()
