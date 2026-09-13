# Copyright 2024-2026 NAEOS Foundation
# SPDX-License-Identifier: Apache-2.0

from fastapi import FastAPI

app = FastAPI(title="web-api-py")

@app.get("/")
def root():
    return {"message": "web-api-py running"}
