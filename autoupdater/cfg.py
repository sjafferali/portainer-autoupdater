import os

PORTAINER_TOKEN = os.getenv("PORTAINER_TOKEN")
PORTAINER_ENDPOINT = os.getenv("PORTAINER_ENDPOINT")
ENABLE_STACK_UPDATE = int(os.getenv("ENABLE_STACK_UPDATE", 0)) == 1
DRY_RUN = int(os.getenv("DRY_RUN", 1)) == 1
LOGLEVEL = os.getenv("LOGLEVEL", "INFO").upper()
RUN_INTERVAL = int(os.getenv("RUN_INTERVAL", 300))
EXCLUDE_STACKS = os.getenv("EXCLUDE_STACKS", "")
INCLUDE_STACKS = os.getenv("INCLUDE_STACKS", "")
