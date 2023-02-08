from cfg import PORTAINER_ENDPOINT
from utils.request import make_request


class Stack:
    def __init__(self, json_data=None):
        self._name = None
        self.env = None
        self.git_config = None
        self.stack_id = None
        self.webhook_id = None

        self.load_json(json_data)

    def refresh(self):
        data = make_request(f"{PORTAINER_ENDPOINT}/api/stacks/{self.stack_id}")
        if data:
            self.load_json(data)

    def load_json(self, data=None):
        if not data:
            return

        self._name = data["Name"]
        self.env = data.get("Env", [])
        self.git_config = data.get("GitConfig")
        self.stack_id = data["Id"]
        self.webhook_id = data["Webhook"]

    @property
    def name(self):
        return self._name

    def stack_file_contents(self):
        data = make_request(f"{PORTAINER_ENDPOINT}/api/stacks/{self.stack_id}/file")
        if data:
            return data.get("StackFileContent")

    def image_status(self):
        data = make_request(f"{PORTAINER_ENDPOINT}/api/stacks/{self.stack_id}/images_status")
        if data and data.get("Status"):
            return data["Status"]

    def update_stack(self):
        if self.git_config:
            return self._git_update_stack()
        return self._file_update_stack()

    def _git_update_stack(self):
        request_body = {
            "env": self.env,
            "prune": True,
            "pullImage": True,
            "repositoryAuthentication": False,
            "repositoryGitCredentialID": 0,
            "repositoryPassword": "",
            "repositoryReferenceName": self.git_config["ReferenceName"],
            "repositoryUsername": ""
        }
        if self.git_config["Authentication"]:
            request_body["repositoryGitCredentialID"] = self.git_config["Authentication"]["GitCredentialID"]
            request_body["repositoryPassword"] = self.git_config["Authentication"]["Password"]
            request_body["repositoryUsername"] = self.git_config["Authentication"]["Username"]

        return make_request(f"{PORTAINER_ENDPOINT}/api/stacks/{self.stack_id}/git/redeploy", "PUT", True, request_body)

    def _file_update_stack(self):
        stack_contents = self.stack_file_contents()
        if not stack_contents:
            print("error getting stack contents")
            return

        request_body = {
            "env": self.env,
            "prune": True,
            "pullImage": True,
            "stackFileContent": self.stack_file_contents(),
            "webhook": self.webhook_id
        }
        return make_request(f"{PORTAINER_ENDPOINT}/api/stacks/{self.stack_id}", "PUT", True, request_body)
