from models.stacks import Stack
from utils.request import make_request
from cfg import PORTAINER_ENDPOINT


def stacks():
    response = make_request(f"{PORTAINER_ENDPOINT}/api/stacks")
    if not response:
        return None

    results = []
    for i in response:
        results.append(Stack(i))
    return results
