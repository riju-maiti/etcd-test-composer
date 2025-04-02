import random, etcd3, string
import numpy as np

# Antithesis SDK
from antithesis.random import (
    random_choice
)

# Antithesis SDK
from antithesis.assertions import (
    unreachable,
)


def put_request(c, key, value):
    try:
        c.put(key, value)
        return True, None
    except Exception as e:
        return False, e

def get_request(c, key):
    try:
        response = c.get(key)
        database_value = response[0].decode('utf-8')
        return True, None, database_value
    except Exception as e:
        return False, e, None

def generate_requests(probabilities, mu, sd):
    num_requests = int(np.random.normal(loc=mu, scale=sd))
    num_requests = max(num_requests, 1)
    request_types = list(probabilities.keys())
    request_weights = probabilities.values()
    return random.SystemRandom().choices(request_types, request_weights, k=num_requests)

def generate_random_string():
    random_str = []
    for _ in range(8):
        random_str.append(random_choice(list(string.ascii_letters + string.digits)))
    return "".join(random_str)

def connect_to_host():
    host = random_choice(["etcd0", "etcd1", "etcd2"])
    try:
        client = etcd3.client(host=host, port=2379)
        print(f"Client: connected to {host}")
        return client
    except Exception as e:
        # Antithesis Assertion: client should always be able to connect to an etcd host
        unreachable("Client failed to connect to an etcd host", {"host":host, "error":e})
        print(f"Client: failed to connect to {host}. exiting")
        sys.exit(1)
