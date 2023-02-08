from cfg import EXCLUDE_STACKS, INCLUDE_STACKS


def stacks(stack_list):
    if not EXCLUDE_STACKS and not INCLUDE_STACKS:
        return stack_list

    excludes = None
    if EXCLUDE_STACKS:
        excludes = EXCLUDE_STACKS.split(",")

    includes = None
    if INCLUDE_STACKS:
        includes = INCLUDE_STACKS.split(",")

    results = []
    for stack in stack_list:
        stack_id = stack.stack_id

        if excludes and str(stack_id) in excludes:
            continue

        if includes and str(stack_id) not in includes:
            continue

        results.append(stack)

    return results
