import requests
import json
import traceback
import logging
from cfg import PORTAINER_TOKEN

logging.getLogger("requests").setLevel(logging.WARNING)
logging.getLogger("urllib3").setLevel(logging.WARNING)


TIMEOUT = 30


def make_request(url, method="GET", auth_header=True, body=None, params=None):
    if body is not None:
        body = json.dumps(body)

    headers = {"Content-Type": "application/json"}
    if auth_header:
        headers["X-API-Key"] = PORTAINER_TOKEN

    try:
        r = requests.request(method, headers=headers, url=url, params=params, data=body, timeout=TIMEOUT, verify=False)
        r.raise_for_status()
    except requests.exceptions.HTTPError as e:
        print("Http Error:", e)
        print(e.response.text)
        return None
    except requests.exceptions.ConnectionError as e:
        print("Error Connecting:", e)
        return None
    except requests.exceptions.Timeout as e:
        print("Timeout Error:", e)
        return None
    except requests.exceptions.RequestException as e:
        print("OOps: Something Else", e)
        return None

    try:
        json_response = r.json()
    except Exception as e:
        print(f"Error occurred converting response to json {e}")
        traceback.print_exc()
        return r.text
    return json_response
