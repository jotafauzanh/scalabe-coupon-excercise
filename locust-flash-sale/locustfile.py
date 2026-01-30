# The :8089 one
# The "Flash Sale" Attack: 50 concurrent requests for a coupon with only 5 items in stock.
# (Result must be exactly 5 claims, 0 remaining).

from locust import HttpUser, task
from gevent.pool import Pool
from requests.adapters import HTTPAdapter
import uuid
import logging

logger = logging.getLogger(__name__)


class FlashSaleUser(HttpUser):
  TOTAL_REQUESTS = 50
  TOTAL_STOCK = 5

  # Full auto, no wait between tasks
  wait_time = lambda self: 0

  def on_start(self):
    # Prereq:
    # - Increase connection pool size
    # - Create 1 unique coupon
    # - Create 50 unique user for each claim attempt

    # Increase urllib3/requests connection pool size (default is 10)
    try:
      adapter = HTTPAdapter(
        pool_connections=self.TOTAL_REQUESTS,
        pool_maxsize=self.TOTAL_REQUESTS,
      )

      self.client.mount("http://", adapter)
      # self.client.mount("https://", adapter)
      logger.info("Increased HTTP connection pool size to %s", self.TOTAL_REQUESTS)
    except Exception as e:
      logger.warning("Failed to bump connection pool size: %s", e)

    # Generate run-specific identifiers for this test run
    self.RUN_ID = str(uuid.uuid4())
    self.COUPON_NAME = f"FLASH_SALE_{self.RUN_ID}"
    self.USER_IDS = [str(uuid.uuid4()) for _ in range(self.TOTAL_REQUESTS)]

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
        logger.info("Coupon %s created (or already existed)", self.COUPON_NAME)
        resp.success()

    # Create 50 users for this run
    for user_id in self.USER_IDS:
      user_body = {
        "name": f"flash_user_{self.RUN_ID}",
        "user_id": user_id,
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
            user_id,
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
        # Print to CLI
        print(message)
        # And to Locust logs
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
  def attack(self):
    # Spawn 50 concurrent claim attempts at once.
    # Each request uses a different user ID.
    # Kill the test after its done, because locust usually just kept going
    pool = Pool(size=self.TOTAL_REQUESTS)

    def claim(user_id: str):
      with self.client.post(
        "/api/coupons/claim",
        json={"user_id": user_id, "coupon_name": self.COUPON_NAME},
        catch_response=True,
      ) as resp:
        if resp.status_code == 200:
          resp.success()
        else:
          resp.failure(f"status={resp.status_code}")

    for user_id in self.USER_IDS:
      pool.spawn(claim, user_id)

    pool.join()

    # Fetch and log final coupon details, so that we dont need to manually check via api/db hit
    self._log_final_coupon_state()

    # Stop test immediately after one attack
    self.environment.runner.quit()
