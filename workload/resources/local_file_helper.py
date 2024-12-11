from filelock import FileLock

def generate_traffic_id():

    with FileLock("/opt/antithesis/local-txt-files/client-traffic-ids.txt.lock"):

        with open("/opt/antithesis/local-txt-files/client-traffic-ids.txt", "r") as f:
            line = f.readline().strip()
        
        if line:
            ids = list(map(int, line.split(',')))
            new_id = ids[-1] + 1
        else:
            ids = []
            new_id = 1
        
        with open("/opt/antithesis/local-txt-files/client-traffic-ids.txt", "w") as f:
            f.write(','.join(map(str, ids + [new_id])))
    
    return new_id

def write_operations(operations):

    with FileLock("/opt/antithesis/local-txt-files/operations.txt.lock"):
        
        with open("/opt/antithesis/local-txt-files/operations.txt", "a") as f:
            for o in operations:
                f.write(f"{o}\n")
