# from locust import HttpUser , task, between

# class AlbumUser(HttpUser):
#     # Wait time between tasks for each simulated user
#     wait_time = between(0, 0.07)

#     @task(3)  # Weighted heavier, because GET is more common
#     def get_albums(self):
#         self.client.get("/albums/99")

#     @task(1)  # Weighted lighter, POST less common
#     def post_album(self):
#         new_album = {
#             "id": "99",
#             "title": "Test Album",
#             "artist": "Locust Bot",
#             "price": 9.99
#         }
#         self.client.post("/albums", json=new_album)
from locust import FastHttpUser, task, between
import random, string

def rand_id(n=6): return "".join(random.choices("0123456789", k=n))

class AlbumUser(FastHttpUser):
    # wait_time = between(0, 0.01)  # smooth CPU; adjust if your laptop gets hot

    def on_start(self):
        # create a unique album per simulated user to avoid 409 conflicts
        self.my_id = rand_id()
        payload = {"id": self.my_id, "title": "Test Album", "artist": "Locust Bot", "price": 9.99}
        self.client.post("/albums", json=payload, name="POST /albums (setup)")

    @task(3)  # 3:1 GET:POST as requested
    def get_hit(self):
        self.client.get(f"/albums/{self.my_id}", name="GET /albums/:id (hit)")

    @task(1)
    def post_album(self):
        # create throwaway IDs to keep writes cheap & avoid contention
        payload = {"id": rand_id(), "title": "T", "artist": "L", "price": 1.23}
        self.client.post("/albums", json=payload, name="POST /albums")
