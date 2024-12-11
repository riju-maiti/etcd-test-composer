#!/usr/bin/env -S python3 -u

# This file serves as the workload's entrypoint. It: 
# 1. Confirms that all nodes in the cluster are available
# 2. Signals "setupComplete" using the Antithesis SDK

import etcd3, time

from antithesis.lifecycle import (
    setup_complete,
)

SLEEP = 10

def check_health():

    node_options = ["etcd0", "etcd1", "etcd2"]

    for i in range(0, len(node_options)):
        #c.get will fail if node is not healthy
        try:
            c = etcd3.client(host=node_options[i], port=2379)
            c.get('setting-up')
            print(f"Workload [first.py]: connection successful with {node_options[i]}")
        except Exception as e:
            print(f"Workload [first.py]: connection failed with {node_options[i]}")
            print(f"Workload [first.py]: error: {e}")
            return False
    return True
    
print("Workload [first.py]: entered")

while True:
    print("Workload [first.py]: checking cluster health")
    if check_health():
        print("Workload [first.py]: cluster is healthy.")
        setup_complete({"Message":"ETCD cluster is healthy"})
        break
    else:
        print(f"Workload [first.py]: cluster is not healthy. retrying in {SLEEP} seconds...")
        time.sleep(SLEEP)

# sleep infinity
time.sleep(31536000)
