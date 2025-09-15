import json

with open('large_maclist.json', 'r') as f:
    data = json.load(f)

# Mask all MAC addresses in the "data" array
if "data" in data:
    data["data"] = ["AA:AA:AA:AA:AA:AA" for _ in data["data"]]

with open('large_maclist_masked.json', 'w') as f:
    json.dump(data, f, indent=4)