# The :8090 one
# The "Double Dip" Attack: 10 concurrent requests from the SAME user for the same coupon.
# (Result must be exactly 1 success, 9 failures).

from locust import HttpUser, task
from gevent.pool import Pool
from requests.adapters import HTTPAdapter
import uuid
import logging

logger = logging.getLogger(__name__)


class DoubleDipUser(HttpUser):
  TOTAL_REQUESTS = 10
  TOTAL_STOCK = 10  # only one should succeed anyway

  # Full auto, no wait between tasks
  wait_time = lambda self: 0

  def on_start(self):
    # Prereq:
    # - Increase connection pool size
    # - Create 1 unique coupon with stock 10
    # - Create 1 user that will attempt to claim it 10 times

    # Increase urllib3/requests connection pool size (default is 10)
    try:
      adapter = HTTPAdapter(
        pool_connections=self.TOTAL_REQUESTS,
        pool_maxsize=self.TOTAL_REQUESTS,
      )

      self.client.mount("http://", adapter)
      # self.client.mount("https://", adapter)
      logger.info(
        "Increased HTTP connection pool size to %s", self.TOTAL_REQUESTS
      )
    except Exception as e:
      logger.warning("Failed to bump connection pool size: %s", e)

    # Generate run-specific identifiers for this test run
    self.RUN_ID = str(uuid.uuid4())
    self.COUPON_NAME = f"DOUBLE_DIP_{self.RUN_ID}"
    self.USER_ID = f"double_dip_user_{self.RUN_ID}"

    # Create coupon for this run
    with self.client.post(
      "/api/coupons",
      json={"name": self.COUPON_NAME, "amount": self.TOTAL_STOCK},
      catch_response=True,
    ) as resp:
      if resp.status_code not in (201, 409):
        logger.error(
          "Failed to create coupon %s, status=%s, body=%s",
          self.COUPON_NAME,
          resp.status_code,
          resp.text,
        )
        resp.failure(f"status={resp.status_code}")
      else:
        logger.info(
          "Coupon %s created (or already existed)", self.COUPON_NAME
        )
        resp.success()

    # Create the single attacker user
    user_body = {
      "name": f"double_dip_user_{self.RUN_ID}",
      "user_id": self.USER_ID,
    }
    with self.client.post(
      "/api/users",
      json=user_body,
      catch_response=True,
    ) as user_resp:
      if user_resp.status_code == 201:
        user_resp.success()
      else:
        # If user already exists or another non-201, log but keep going
        logger.warning(
          "User create failed for %s, status=%s, body=%s",
          self.USER_ID,
          user_resp.status_code,
          user_resp.text,
        )
        user_resp.failure(f"status={user_resp.status_code}")

  # Return coupon details so we dont need to manually hit the api
  def _log_final_coupon_state(self):
    with self.client.get(
      f"/api/coupons/{self.COUPON_NAME}", catch_response=True
    ) as resp:
      if resp.status_code == 200:
        details = resp.json()
        message = f"Final coupon state for {self.COUPON_NAME}: {details}"
        print(message)
        logger.info(message)
        resp.success()
      else:
        logger.error(
          "Failed to fetch coupon %s, status=%s, body=%s",
          self.COUPON_NAME,
          resp.status_code,
          resp.text,
        )
        resp.failure(f"status={resp.status_code}")

  @task
  def double_dip(self):
    pool = Pool(size=self.TOTAL_REQUESTS)
    results = {"success": 0, "failure": 0}

    def claim():
      with self.client.post(
        "/api/coupons/claim",
        json={
          "user_id": self.USER_ID,
          "coupon_name": self.COUPON_NAME,
        },
        catch_response=True,
      ) as resp:
        if resp.status_code == 200:
          results["success"] += 1
          resp.success()
        else:
          results["failure"] += 1
          resp.failure(f"status={resp.status_code}")

    # Spawn 10 concurrent claim attempts from the same user
    for _ in range(self.TOTAL_REQUESTS):
      pool.spawn(claim)

    pool.join()

    # Fetch and log final coupon details
    self._log_final_coupon_state()

    # Stop test immediately after one attack
    self.environment.runner.quit()
