# Copyright 2025 NAEOS contributors
# SPDX-License-Identifier: Apache-2.0

from fastapi import FastAPI

app = FastAPI(title="web-api-py")

@app.get("/")
def root():
    return {"message": "web-api-py running"}
