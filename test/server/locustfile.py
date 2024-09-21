from locust import HttpUser, TaskSet, task, between

class UserBehavior(TaskSet):
    
    @task(1)
    def get_homepage(self):
        self.client.get("/")
    
    @task(2)
    def get_about_page(self):
        self.client.get("/about")

class WebsiteUser(HttpUser):
    tasks = [UserBehavior]
    wait_time = between(1, 3)
    host = "http://127.0.0.1:8080"  # Specify the full URL schema
