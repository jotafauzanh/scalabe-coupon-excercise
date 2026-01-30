# The :8091 one
from locust import HttpUser, task, between

# HelloWorld!
class QuickstartUser(HttpUser):
    wait_time = between(0, 1)

    # def on_start(self):
    #     self.client.get("/health")

    @task
    def health(self):
      print("executing health")
      self.client.get("/api/health")
