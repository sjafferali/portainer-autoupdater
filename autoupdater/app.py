import logging
import time

from utils import portainer
from utils import filters
from cfg import LOGLEVEL, DRY_RUN, ENABLE_STACK_UPDATE, RUN_INTERVAL

logging.basicConfig(format='%(asctime)s %(levelname)s %(message)s')
logger = logging.getLogger()
logger.setLevel(LOGLEVEL)


def update_stacks(dry=False):
    stacks = portainer.stacks()
    if not stacks:
        logger.error("failed to get stacks")
        return

    filtered_stacks = filters.stacks(stacks)
    logger.debug("found %d stacks to check", len(filtered_stacks))
    for stack in filtered_stacks:
        image_status = stack.image_status()
        if not image_status:
            logger.error("failed to get image status for stack %s/%d", stack.name, stack.stack_id)
            continue

        if image_status == "updated":
            logger.debug("image for stack %s/%d is up to date", stack.name, stack.stack_id)
            continue

        logger.info("image for stack %s/%d is outdated", stack.name, stack.stack_id)
        if not dry:
            logger.info("performing update for stack %s/%d", stack.name, stack.stack_id)
            stack.update_stack()


def main():
    if DRY_RUN:
        logger.warning("DRY_RUN passed. No updates will be performed.")

    if ENABLE_STACK_UPDATE:
        logger.debug("performing stack updates run")
        update_stacks(DRY_RUN)


if __name__ == "__main__":
    while True:
        main()
        logger.debug("sleeping for %d seconds", RUN_INTERVAL)
        time.sleep(RUN_INTERVAL)
