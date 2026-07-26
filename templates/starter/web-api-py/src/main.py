from fastapi import FastAPI

app = FastAPI(title="web-api-py")

@app.get("/")
def root():
    return {"message": "web-api-py running"}
