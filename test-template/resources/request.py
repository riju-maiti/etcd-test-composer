import time, random
import numpy as np

def put_request(c, key, value):
    start = time.monotonic_ns()
    try:
        response = c.put(key, value)
        end = time.monotonic_ns()
        return True, start, end, response, None
    except Exception as e:
        end = time.monotonic_ns()
        return False, start, end, None, e

def get_request(c, key):
    start = time.monotonic_ns()
    try:
        response = c.get(key)
        end = time.monotonic_ns()
        return True, start, end, response, None
    except Exception as e:
        end = time.monotonic_ns()
        return False, start, end, None, e 

def generate_requests(mean, sd, probabilities):
    num_requests = int(np.random.normal(loc=mean, scale=sd))
    num_requests = max(num_requests, 1)
    request_types = list(probabilities.keys())
    request_weights = probabilities.values()
    return random.SystemRandom().choices(request_types, request_weights, k=num_requests)
